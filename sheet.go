package sheet

import (
	"errors"
	"reflect"
	"strings"
	"time"
	"unicode"
)

const (
	tagName = "sheet"
)

var (
	typeOfTime = reflect.TypeOf(time.Time{})
)

func Marshal(v interface{}) ([][]interface{}, error) {
	return NewEncoder().Encode(v)
}

func Unmarshal(formats [][]string, values [][]string, v interface{}) error {
	return NewDecoder(formats).Decode(values, v)
}

type HeaderOption struct {
	Title bool
}

func Header(v interface{}, opt *HeaderOption) ([][]interface{}, error) {
	cells, err := getHeader(v)
	if err != nil {
		return nil, err
	}
	maxColumn := 0
	maxRow := 0
	for _, cell := range cells {
		if maxColumn < cell.column {
			maxColumn = cell.column
		}
		if maxRow < cell.row {
			maxRow = cell.row
		}
	}
	if opt != nil && opt.Title {
		maxRow++
	}
	tables := make([][]interface{}, maxRow+1)
	for i := range tables {
		tables[i] = make([]interface{}, maxColumn+1)
	}
	for _, cell := range cells {
		tables[cell.row][cell.column] = cell.key
		if opt != nil && opt.Title {
			tables[maxRow+1][cell.column] = cell.title
		}
	}

	return tables, err
}

type headerCells []headerCell

func getHeader(v interface{}) (headerCells, error) {
	rt, rv := reflect.TypeOf(v), reflect.ValueOf(v)

	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		return getHeader(rv.Elem().Interface())
	}

	if rv.Kind() != reflect.Struct && rv.Kind() != reflect.Interface {
		return nil, errors.New("")
	}

	titles := map[int]string{}
	column := 0
	row := 0
	headers := headerCells{}
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if !unicode.IsUpper(rune(field.Name[0])) {
			continue
		}
		tag := field.Tag.Get(tagName)
		if tag == "-" {
			continue
		}
		key := field.Name
		tags := strings.Split(tag, ",")
		if len(tags) > 0 && tags[0] != "" {
			key = tags[0]
		}
		opt := newOption(tag)
		if opt.isDatetime {
			key += ":datetime"
		}
		headers = append(headers, headerCell{
			column: column,
			row:    row,
			key:    key,
			title:  opt.title,
		})
		titles[column] = opt.title

		value := rv.Field(i)
		if value.Kind() == reflect.Ptr && !value.IsNil() {
			value = value.Elem()
		}
		val := header(value)
		if len(val) > 0 {
			maxColumn := column
			for _, cel := range val {
				col := column + cel.column
				headers = append(headers, headerCell{
					column: col,
					row:    row + cel.row + 1,
					key:    cel.key,
					title:  cel.title,
				})
				if maxColumn < col {
					maxColumn = col
				}
				titles[col] = cel.title
			}
			column = maxColumn + 1
			continue
		}
		column++
	}

	return headers, nil
}

func header(v reflect.Value) headerCells {
	switch v.Kind() {
	case reflect.Struct:
		headers, err := getHeader(v.Interface())
		if err != nil {
			panic(err)
		}
		return headers
	case reflect.Array, reflect.Slice:
		rt := reflect.TypeOf(v.Interface())
		rv := reflect.MakeSlice(rt, 1, 1)
		return header(rv.Index(0))
	}
	return nil
}
