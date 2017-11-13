package sheet

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var rowsPool = sync.Pool{
	New: func() interface{} {
		return &rows{
			list: make([]int, 0, 10),
		}
	},
}

func getRowsPool() *rows {
	return rowsPool.Get().(*rows)
}

func resetRowsPool(rows *rows) {
	rows.truncate()
	rowsPool.Put(rows)
}

type rows struct {
	list []int
}

func (r *rows) add(idx int) {
	r.list = append(r.list, idx)
}

func (r *rows) length() int {
	return len(r.list)
}

func (r *rows) truncate() {
	r.list = r.list[:0]
}

type decoder struct {
	formats [][]string
	values  [][]string
}

func newDecoder(formats [][]string) *decoder {
	dec := &decoder{}
	dec.setFormat(formats)
	return dec
}

func (dec *decoder) setFormat(formats [][]string) {
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

func (dec *decoder) Decode(values [][]string, v interface{}) error {
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
		keyIdx := strings.Index(key, ":")
		var opt *option
		if keyIdx > 0 && keyIdx+1 < len(key) {
			// option
			opt = newOption(key[keyIdx+1:], false)
		}
		if keyIdx > 0 {
			key = key[:keyIdx]
		}
		field, ok := rv.Type().FieldByName(key)
		if ok && field.Tag.Get(tagName) != "-" {
			value := rv.FieldByName(key)
			if value.IsValid() {
				dec.decode(value, row, column, opt)
			}
		}
		resetOption(opt)
	}
	return nil
}

func (dec *decoder) decode(v reflect.Value, row, column int, opt *option) error {
	switch v.Kind() {
	case reflect.Ptr:
		elem := reflect.New(v.Type().Elem())
		switch elem.Elem().Kind() {
		case reflect.Struct:
			isExist := false
			for i, format := range dec.formats[row+1][column:] {
				if format == "" {
					break
				}
				if x := dec.getValue(row, column+i); x != "" {
					isExist = true
					break
				}
			}
			if isExist {
				if err := dec.decode(elem.Elem(), row, column, opt); err != nil {
					return err
				}
				v.Set(elem)
			}
		default:
			if x := dec.getValue(row, column); x != "" {
				if err := dec.decode(elem.Elem(), row, column, opt); err != nil {
					return err
				}
				v.Set(elem)
			}
		}
	case reflect.Struct:
		switch v.Type() {
		case typeOfTime:
			x := dec.getValue(row, column)
			t, err := decodeDatetime(x, opt)
			if err != nil {
				return err
			}
			v.Set(reflect.ValueOf(t))
		default:
			if err := dec.decodeStruct(v, row, column, 0); err != nil {
				return err
			}
		}
	case reflect.Array:
		switch v.Index(0).Kind() {
		case reflect.Ptr:
			pType := reflect.New(v.Index(0).Type().Elem())
			switch pType.Elem().Kind() {
			case reflect.Struct:
				rows := dec.targetRows(row, column)
				for idx, i := range rows.list {
					if idx >= v.Len() {
						break
					}
					isExist := false
					for j, format := range dec.formats[row+1][column:] {
						if format == "" {
							break
						}
						if format == "_index" {
							continue
						}
						if x := dec.getValue(row+i, column+j); x != "" {
							isExist = true
							break
						}
					}
					if !isExist {
						continue
					}
					elem := reflect.New(pType.Type().Elem())
					if err := dec.decodeStruct(elem.Elem(), row, column, i); err != nil {
						return err
					}
					v.Index(i).Set(elem)
				}
				resetRowsPool(rows)
			default:
				for i := 0; i < v.Len(); i++ {
					x := dec.getValue(row+i, column)
					if x == "" {
						continue
					}
					elem := reflect.New(v.Index(i).Type().Elem())
					if err := dec.set(elem.Elem(), x, opt); err != nil {
						return err
					}
					v.Index(i).Set(elem)
				}
			}
		case reflect.Struct:
			rows := dec.targetRows(row, column)
			for _, i := range rows.list {
				if err := dec.decodeStruct(v.Index(i), row, column, i); err != nil {
					return err
				}
			}
			resetRowsPool(rows)
		default:
			for i := 0; i < v.Len(); i++ {
				x := dec.getValue(row+i, column)
				if err := dec.set(v.Index(i), x, opt); err != nil {
					return err
				}
			}
		}
	case reflect.Slice:
		elems := reflect.MakeSlice(v.Type(), 0, 1) // 最終的に蓄積するスライス
		rv := reflect.MakeSlice(v.Type(), 1, 1).Index(0)
		switch rv.Kind() {
		case reflect.Ptr:
			switch rv.Type().Elem().Kind() {
			case reflect.Struct:
				rows := dec.targetRows(row, column)
				for _, i := range rows.list {
					isExist := false
					for j, format := range dec.formats[row+1][column:] {
						if format == "" {
							break
						}
						if format == "_index" {
							continue
						}
						if x := dec.getValue(row+i, column+j); x != "" {
							isExist = true
							break
						}
					}
					if isExist {
						elem := reflect.New(rv.Type().Elem())
						if err := dec.decodeStruct(elem.Elem(), row, column, i); err != nil {
							return err
						}
						elems = reflect.Append(elems, elem)
					} else {
						elems = reflect.Append(elems, rv)
					}
				}
				resetRowsPool(rows)
			default:
				if opt != nil && opt.isCSV {
					for _, x := range strings.Split(dec.getValue(row, column), ",") {
						elem := reflect.New(rv.Type().Elem())
						if err := dec.set(elem.Elem(), x, opt); err != nil {
							return err
						}
						elems = reflect.Append(elems, elem)
					}
				} else {
					rows := dec.targetRows(row, column)
					if rows.length() != 0 {
						size := rows.list[rows.length()-1]
						for i := 0; i <= size; i++ {
							x := dec.getValue(row+i, column)
							if x == "" {
								elems = reflect.Append(elems, rv)
								continue
							}
							elem := reflect.New(rv.Type().Elem())
							if err := dec.set(elem.Elem(), x, opt); err != nil {
								return err
							}
							elems = reflect.Append(elems, elem)
						}
					}
					resetRowsPool(rows)
				}
			}
		case reflect.Struct:
			rows := dec.targetRows(row, column)
			for _, i := range rows.list {
				elem := reflect.New(rv.Type()).Elem()
				if err := dec.decodeStruct(elem, row, column, i); err != nil {
					return err
				}
				elems = reflect.Append(elems, elem)
			}
			resetRowsPool(rows)
		default:
			if opt != nil && opt.isCSV {
				for _, x := range strings.Split(dec.getValue(row, column), ",") {
					elem := reflect.New(rv.Type()).Elem()
					if err := dec.set(elem, x, opt); err != nil {
						return err
					}
					elems = reflect.Append(elems, elem)
				}
			} else {
				rows := dec.targetRows(row, column)
				if rows.length() != 0 {
					size := rows.list[rows.length()-1]
					for i := 0; i <= size; i++ {
						x := dec.getValue(row+i, column)
						elem := reflect.New(rv.Type()).Elem()
						if err := dec.set(elem, x, opt); err != nil {
							return err
						}
						elems = reflect.Append(elems, elem)
					}
				}
				resetRowsPool(rows)
			}
		}
		v.Set(elems)
	default:
		x := dec.getValue(row, column)
		if err := dec.set(v, x, opt); err != nil {
			return err
		}
	}
	return nil
}

func (dec *decoder) decodeStruct(v reflect.Value, row, column, idx int) error {
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
		keyIdx := strings.Index(key, ":")
		var opt *option
		if keyIdx > 0 && keyIdx+1 < len(key) {
			// option
			opt = newOption(key[keyIdx+1:], false)
		}
		if keyIdx > 0 {
			key = key[:keyIdx]
		}
		field, ok := v.Type().FieldByName(key)
		if ok && field.Tag.Get(tagName) != "-" {
			elem := v.FieldByName(key)
			if !elem.IsValid() {
				continue
			}
			if err := dec.decode(elem, row+idx, column+i, opt); err != nil {
				return err
			}
		}
		resetOption(opt)
	}
	return nil
}

func (dec *decoder) targetRows(row, column int) *rows {
	rows := getRowsPool()
	for i := 0; i < len(dec.values); i++ {
		if x := dec.getValue(row+i, column); x != "" {
			rows.add(i)
		}
	}
	return rows
}

func (dec *decoder) getValue(row, column int) string {
	if row < len(dec.values) && column < len(dec.values[row]) {
		return dec.values[row][column]
	}
	return ""
}

func (dec *decoder) set(v reflect.Value, value string, opt *option) error {
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
		if opt != nil && opt.isDatetime {
			t, err := decodeDatetime(value, opt)
			if err != nil {
				return err
			}
			value = strconv.FormatInt(t.Unix(), 10)
		}
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
