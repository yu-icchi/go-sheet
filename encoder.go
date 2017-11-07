package sheet

import (
	"reflect"
	"sync"
	"time"
	"unicode"
)

var cellsPool = sync.Pool{
	New: func() interface{} {
		return &cells{
			list: make([]cell, 0, 1024),
		}
	},
}

func newCellPool() *cells {
	return cellsPool.Get().(*cells)
}

func resetCellPool(cells *cells) {
	cells.truncate()
	cellsPool.Put(cells)
}

type cell struct {
	column int
	row    int
	value  interface{}
}

type cells struct {
	list []cell
}

func (c *cells) add(cell cell) {
	c.list = append(c.list, cell)
}

func (c *cells) truncate() {
	c.list = c.list[:0]
}

type Encoder struct {
	cells     *cells
	maxColumn int
	maxRow    int
}

func NewEncoder() *Encoder {
	return &Encoder{
		maxColumn: 0,
		maxRow:    0,
	}
}

func (enc *Encoder) init() {
	enc.cells = newCellPool()
	enc.maxColumn = 0
	enc.maxRow = 0
}

func (enc *Encoder) reset() {
	resetCellPool(enc.cells)
}

func (enc *Encoder) Encode(v interface{}) ([][]interface{}, error) {
	enc.init()
	defer enc.reset()
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}
	if _, err := enc.reflectStruct(rv, 0, 0, false); err != nil {
		return nil, err
	}
	values := make([][]interface{}, enc.maxRow+1)
	for i := range values {
		values[i] = make([]interface{}, enc.maxColumn+1)
	}
	for _, cell := range enc.cells.list {
		values[cell.row][cell.column] = cell.value
	}
	return values, nil
}

func (enc *Encoder) reflectStruct(v reflect.Value, column, row int, isNil bool) (int, error) {
	n := 0
	for i := 0; i < v.Type().NumField(); i++ {
		field := v.Type().Field(i)
		if !unicode.IsUpper(rune(field.Name[0])) {
			continue
		}
		tag := field.Tag.Get(tagName)
		if tag == "-" {
			continue
		}
		opt := newOption(tag)
		addNum, err := enc.reflectValue(v.Field(i), column+n, row, opt, isNil)
		if err != nil {
			return 0, err
		}
		if addNum > 0 {
			n += addNum
		} else {
			n++
		}
		resetOption(opt)
	}
	return n, nil
}

func (enc *Encoder) reflectList(v reflect.Value, isStruct bool, column int, opt *option, isNil bool) (int, error) {
	col := 0
	for i := 0; i < v.Len(); i++ {
		n := 0
		if isStruct {
			enc.add(i+1, column, i)
			n = 1
		}
		n, err := enc.reflectValue(v.Index(i), column+n, i, opt, isNil)
		if err != nil {
			return 0, err
		}
		if col < n {
			col = n
		}
	}
	return col, nil
}

func (enc *Encoder) reflectValue(v reflect.Value, column, row int, opt *option, isNil bool) (int, error) {
	switch v.Kind() {
	case reflect.Ptr:
		isNil = v.IsNil()
		if isNil {
			v = reflect.New(v.Type().Elem())
		}
		n, err := enc.reflectValue(v.Elem(), column, row, opt, isNil)
		if err != nil {
			return 0, err
		}
		return n, nil
	case reflect.Struct:
		switch v.Type() {
		case typeOfTime:
			if opt != nil && opt.isDatetime {
				t, err := encodeDatetime(v)
				if err != nil {
					return 0, err
				}
				enc.add(t, column, row)
			} else {
				val := v.Interface().(time.Time)
				txt, err := val.MarshalText()
				if err != nil {
					return 0, err
				}
				enc.add(string(txt), column, row)
			}
		default:
			n, err := enc.reflectStruct(v, column, row, isNil)
			if err != nil {
				return 0, err
			}
			return n, nil
		}
	case reflect.Array:
		rv := reflect.New(v.Type()).Elem().Index(0)
		isStruct := rv.Kind() == reflect.Struct
		col, err := enc.reflectList(v, isStruct, column, opt, isNil)
		if err != nil {
			return 0, err
		}
		if isStruct {
			col++
		}
		return col, nil
	case reflect.Slice:
		col := 0
		rv := reflect.MakeSlice(v.Type(), 1, 1).Index(0)
		isStruct := rv.Kind() == reflect.Struct
		if v.Len() > 0 {
			var err error
			col, err = enc.reflectList(v, isStruct, column, opt, isNil)
			if err != nil {
				return 0, err
			}
		} else {
			n := 0
			if isStruct {
				enc.add(0, column, row)
				n = 1
			}
			n, err := enc.reflectValue(rv, column+n, row, opt, isNil)
			if err != nil {
				return 0, err
			}
			if col < n {
				col = n
			}
		}
		if isStruct {
			col++
		}
		return col, nil
	}
	if opt != nil && opt.isDatetime {
		t, err := encodeDatetime(v)
		if err != nil {
			return 0, err
		}
		enc.add(t, column, row)
	} else if isNil {
		enc.add(nil, column, row)
	} else {
		switch v.Kind() {
		case reflect.String:
			enc.add(v.String(), column, row)
		case reflect.Int:
			enc.add(int(v.Int()), column, row)
		case reflect.Int8:
			enc.add(int8(v.Int()), column, row)
		case reflect.Int16:
			enc.add(int16(v.Int()), column, row)
		case reflect.Int32:
			enc.add(int32(v.Int()), column, row)
		case reflect.Int64:
			enc.add(v.Int(), column, row)
		case reflect.Uint:
			enc.add(uint(v.Uint()), column, row)
		case reflect.Uint8:
			enc.add(uint8(v.Uint()), column, row)
		case reflect.Uint16:
			enc.add(uint16(v.Uint()), column, row)
		case reflect.Uint32:
			enc.add(uint32(v.Uint()), column, row)
		case reflect.Uint64:
			enc.add(v.Uint(), column, row)
		case reflect.Float32:
			enc.add(float32(v.Float()), column, row)
		case reflect.Float64:
			enc.add(v.Float(), column, row)
		case reflect.Bool:
			enc.add(v.Bool(), column, row)
		}
	}
	return 0, nil
}

func (enc *Encoder) add(v interface{}, column, row int) {
	enc.cells.add(cell{
		column: column,
		row:    row,
		value:  v,
	})
	if enc.maxColumn < column {
		enc.maxColumn = column
	}
	if enc.maxRow < row {
		enc.maxRow = row
	}
}
