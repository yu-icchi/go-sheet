package sheet

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	tagName = "sheet"
)

func Marshal(v interface{}) ([][]interface{}, error) {
	cells, err := getCells(v, false)
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

	// sync.poolが使えそう
	tables := make([][]interface{}, maxRow+1)
	for i := range tables {
		tables[i] = make([]interface{}, maxColumn+1)
	}
	for _, cell := range cells {
		tables[cell.row][cell.column] = cell.value
	}

	return tables, err
}

func Unmarshal(formats [][]string, values [][]string, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	row := 0
	for column := range formats[row] {
		name := formats[row][column]
		if name == "" {
			continue
		}
		v := rv.FieldByName(name)
		if err := _unmarshal(v, formats, row, column, values, column, 0); err != nil {
			return err
		}
	}
	return nil
}

func _unmarshal(v reflect.Value, formats [][]string, row, column int, values [][]string, originColumn, offset int) error {
	switch v.Kind() {
	case reflect.Ptr:
		elem := reflect.New(v.Type().Elem())
		if err := _unmarshal(elem.Elem(), formats, row, column, values, originColumn, offset); err != nil {
			return err
		}
		v.Set(elem)
	case reflect.Struct:
		if err := _unmarshalStruct(v, formats, row, column, values, 0); err != nil {
			return err
		}
	case reflect.Slice:
		sliceType := reflect.MakeSlice(v.Type(), 1, 1).Index(0) // 1件分の要素を作成し内容を取得する
		elems := reflect.MakeSlice(v.Type(), 0, 1)              // 最終的に蓄積するスライス
		switch sliceType.Kind() {
		case reflect.Ptr:
			// struct, sliceのチェックが必要そう
			pType := reflect.New(sliceType.Type().Elem())
			switch pType.Elem().Kind() {
			case reflect.Struct:
				rows := []int{}
				for i := 0; i < len(values); i++ {
					if values[row+i][column] != "" {
						rows = append(rows, i)
					}
				}
				for _, i := range rows {
					vv := reflect.New(sliceType.Type().Elem())
					if err := _unmarshalStruct(vv.Elem(), formats, row, column, values, i); err != nil {
						return err
					}
					elems = reflect.Append(elems, vv)
				}
			default:
				for i := offset; i < len(values); i++ {
					x := values[row+i][column] // todo...index out of rangeになるのでチェックが必要
					if x == "" {
						break
					}
					vv := reflect.New(sliceType.Type().Elem())
					if err := setValue(vv.Elem(), x); err != nil {
						return err
					}
					elems = reflect.Append(elems, vv)
				}
			}
		case reflect.Struct:
			rows := []int{}
			for i := 0; i < len(values); i++ {
				if values[row+i][column] != "" {
					rows = append(rows, i)
				}
			}
			for _, i := range rows {
				elem := reflect.New(sliceType.Type()).Elem()
				if err := _unmarshalStruct(elem, formats, row, column, values, i); err != nil {
					return err
				}
				elems = reflect.Append(elems, elem)
			}
		default:
			for i := offset; i < len(values); i++ {
				x := values[row+i][column]
				if x == "" {
					break
				}
				if column != originColumn && i-offset > 0 && values[row+i][originColumn] != "" {
					break
				}
				elem := reflect.New(sliceType.Type()).Elem()
				if err := setValue(elem, x); err != nil {
					return err
				}
				elems = reflect.Append(elems, elem)
			}
		}
		v.Set(elems)
	case reflect.Array:
		return errors.New("unsupported array")
	default:
		value := values[row][column]
		if err := setValue(v, value); err != nil {
			return err
		}
	}
	return nil
}

func _unmarshalStruct(v reflect.Value, formats [][]string, row, column int, values [][]string, offset int) error {
	l := 1
	for _, format := range formats[row][column+1:] {
		if format != "" {
			break
		}
		l++
	}
	for i, name := range formats[row+1][column : column+l] {
		if name == "" {
			break
		}
		value := values[row+offset][column+i]
		if value == "" {
			break
		}
		elem := v.FieldByName(name)
		if elem.Kind() == reflect.Slice {
			_unmarshal(elem, formats, row, column+i, values, column, offset)
		}
		if err := setValue(elem, value); err != nil {
			return err
		}
	}
	return nil
}

func setValue(v reflect.Value, value string) error {
	switch v.Kind() {
	case reflect.Bool:
		x, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		v.SetBool(x)
	case reflect.String:
		v.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		x, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(x)
	case reflect.Float32, reflect.Float64:
		x, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		v.SetFloat(x)
	}
	return nil
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

type headerCell struct {
	column int
	row    int
	key    string
	title  string
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
		opt := genOption(tag)
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

type cell struct {
	column int
	row    int
	value  interface{}
}

type cells []cell

func getCells(v interface{}, isNil bool) (cells, error) {
	rv := reflect.ValueOf(v)

	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		return getCells(rv.Elem().Interface(), isNil)
	}

	if rv.Kind() != reflect.Struct && rv.Kind() != reflect.Interface {
		return nil, errors.New("")
	}

	column := 0
	row := 0
	list := cells{}
	for i := 0; i < rv.Type().NumField(); i++ {
		field := rv.Type().Field(i)
		if !unicode.IsUpper(rune(field.Name[0])) {
			continue
		}
		tag := field.Tag.Get(tagName)
		if tag == "-" {
			continue
		}

		value := rv.Field(i)
		if value.Kind() == reflect.Ptr && !value.IsNil() {
			value = value.Elem()
		}

		option := genOption(tag)
		data, err := getValue(value, isNil, option)
		if err != nil {
			return nil, err
		}
		switch data.(type) {
		case cells:
			for _, cel := range data.(cells) {
				list = append(list, cell{
					column: column + cel.column,
					row:    cel.row,
					value:  cel.value,
				})
			}
			column += len(data.(cells))
		case []interface{}:
			maxColumn := column
			for i, v := range data.([]interface{}) {
				if cel, ok := v.(cells); ok {
					maxRow := 0
					for _, c := range cel {
						col := column + c.column
						list = append(list, cell{
							column: col,
							row:    row + c.row + i,
							value:  c.value,
						})
						if maxRow < c.row {
							maxRow = c.row
						}
						if maxColumn < col {
							maxColumn = col
						}
					}
					row += maxRow
				} else {
					list = append(list, cell{
						column: column,
						row:    row + i,
						value:  v,
					})
				}
			}
			column = maxColumn + 1
		default:
			list = append(list, cell{
				column: column,
				row:    row,
				value:  data,
			})
			column++
		}
	}
	return list, nil
}

func getValue(v reflect.Value, isNil bool, option *option) (interface{}, error) {
	switch v.Interface().(type) {
	case time.Time:
		if option.isDatetime {
			return encodeDatetime(v)
		}
		val, ok := v.Interface().(time.Time)
		if !ok {
			return nil, errors.New("mismatch")
		}
		txt, err := val.MarshalText()
		if err != nil {
			return nil, err
		}
		return string(txt), nil
	default:
		kind := v.Kind()
		switch kind {
		case reflect.Struct:
			return getCells(v.Interface(), isNil)
		case reflect.Map:
			return nil, errors.New("unsupported map")
		case reflect.Slice, reflect.Array:
			l := v.Len()
			list := make([]interface{}, 0, l)
			if l > 0 {
				for i := 0; i < l; i++ {
					ret, err := getValue(v.Index(i), false, option)
					if err != nil {
						return nil, err
					}
					list = append(list, ret)
				}
			} else {
				// 実際の値が無い場合はvalue:nilの状態でカラムを作成する
				rt := reflect.TypeOf(v.Interface())
				rv := reflect.MakeSlice(rt, 1, 1)
				ret, err := getValue(rv.Index(0), true, option)
				if err != nil {
					return nil, err
				}
				list = append(list, ret)
			}
			return list, nil
		default:
			if isNil {
				return nil, nil
			}

			// Datetimeオプションが付いていれば形式を変更する
			if option != nil && option.isDatetime {
				return encodeDatetime(v)
			}

			switch kind {
			case reflect.String:
				return v.String(), nil
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return v.Int(), nil
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return v.Uint(), nil
			case reflect.Float32, reflect.Float64:
				return v.Float(), nil
			case reflect.Bool:
				return v.Bool(), nil
			default:
				return v.Interface(), nil
			}
		}
	}
	return nil, nil
}
