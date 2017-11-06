package sheet

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

type Decoder struct {
	formats [][]string
	values  [][]string
}

func NewDecoder(formats [][]string) *Decoder {
	dec := &Decoder{}
	dec.setFormat(formats)
	return dec
}

func (dec *Decoder) setFormat(formats [][]string) {
	maxColumn := 0
	for i := range formats {
		if maxColumn < len(formats[i]) {
			maxColumn = len(formats[i])
		}
	}
	ret := make([][]string, len(formats))
	for i := range ret {
		ret[i] = make([]string, maxColumn)
	}
	for i := range formats {
		for j := range formats[i] {
			ret[i][j] = formats[i][j]
		}
	}
	dec.formats = ret
}

func (dec *Decoder) Decode(values [][]string, v interface{}) error {
	dec.values = values

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("invalid decode error")
	}

	rv = rv.Elem()
	row := 0
	for column := range dec.formats[row] {
		key := dec.formats[row][column]
		if key == "" {
			continue
		}
		value := rv.FieldByName(key)
		dec.decode(value, row, column)
	}
	return nil
}

func (dec *Decoder) decode(v reflect.Value, row, column int) error {
	switch v.Interface().(type) {
	case time.Time:
		x := dec.getValue(row, column)
		fmt.Println("====>", v, x)
	default:
		switch v.Kind() {
		case reflect.Ptr:
			// todo...nilの場合はどうするかな。。。
			elem := reflect.New(v.Type().Elem())
			if err := dec.decode(elem.Elem(), row, column); err != nil {
				return err
			}
			v.Set(elem)
		case reflect.Struct:
			if err := dec.decodeStruct(v, row, column, 0); err != nil {
				return err
			}
		case reflect.Array:
			return errors.New("unsupported array")
		case reflect.Slice:
			elems := reflect.MakeSlice(v.Type(), 0, 1) // 最終的に蓄積するスライス
			rv := reflect.MakeSlice(v.Type(), 1, 1).Index(0)
			switch rv.Kind() {
			case reflect.Ptr:
				pType := reflect.New(rv.Type().Elem())
				switch pType.Elem().Kind() {
				case reflect.Struct:
					rows := dec.targetRows(row, column)
					for _, i := range rows {
						elem := reflect.New(rv.Type().Elem())
						if err := dec.decodeStruct(elem.Elem(), row, column, i); err != nil {
							return err
						}
						elems = reflect.Append(elems, elem)
					}
				default:
					for i := 0; i < len(dec.values); i++ {
						x := dec.getValue(row+i, column)
						elem := reflect.New(rv.Type().Elem())
						if err := dec.set(elem.Elem(), x); err != nil {
							return err
						}
						elems = reflect.Append(elems, elem)
					}
				}
			case reflect.Struct:
				rows := dec.targetRows(row, column)
				fmt.Println("-->", rows)
				for _, i := range rows {
					elem := reflect.New(rv.Type()).Elem()
					if err := dec.decodeStruct(elem, row, column, i); err != nil {
						return err
					}
					fmt.Println("-->", elem)
					elems = reflect.Append(elems, elem)
				}
			default:
				for i := 0; i < len(dec.values); i++ {
					x := dec.getValue(row+i, column)
					elem := reflect.New(rv.Type()).Elem()
					if err := dec.set(elem, x); err != nil {
						return err
					}
					elems = reflect.Append(elems, elem)
				}
			}
			v.Set(elems)
		default:
			x := dec.getValue(row, column)
			if err := dec.set(v, x); err != nil {
				return err
			}
		}
	}
	return nil
}

func (dec *Decoder) decodeStruct(v reflect.Value, row, column, idx int) error {
	l := 1
	for _, format := range dec.formats[row][column+1:] {
		if format != "" {
			break
		}
		l++
	}
	for i, key := range dec.formats[row+1][column : column+l] {
		if key == "" {
			break
		}
		elem := v.FieldByName(key)
		if !elem.IsValid() {
			continue
		}
		if err := dec.decode(elem, row+idx, column+i); err != nil {
			return err
		}
	}
	return nil
}

func (dec *Decoder) targetRows(row, column int) []int {
	rows := make([]int, 0, len(dec.values))
	for i := 0; i < len(dec.values); i++ {
		if x := dec.getValue(row+i, column); x != "" {
			rows = append(rows, i)
		}
	}
	return rows
}

func (dec *Decoder) getValue(row, column int) string {
	if row < len(dec.values) && column < len(dec.values[row]) {
		return dec.values[row][column]
	}
	return ""
}

func (dec *Decoder) set(v reflect.Value, value string) error {
	if value == "" {
		return nil
	}
	switch v.Kind() {
	case reflect.String:
		v.SetString(value)
	case reflect.Bool:
		x, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		v.SetBool(x)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		x, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(x)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		x, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(x)
	case reflect.Float32, reflect.Float64:
		x, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		v.SetFloat(x)
	}
	return nil
}
