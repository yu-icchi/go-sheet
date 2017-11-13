package sheet

import (
	"fmt"
	"reflect"
	"unicode"
)

type headerCell struct {
	column int
	row    int
	key    string
	title  string
}

type headerEncoder struct {
	cells []headerCell
}

func newHeaderEncoder() *headerEncoder {
	return &headerEncoder{
		cells: []headerCell{},
	}
}

func (enc *headerEncoder) Encode(v interface{}) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}

	enc.encode(rv)
}

func (enc *headerEncoder) encode(v reflect.Value) {
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
		key := field.Name
		opt := newOption(tag, true)
		if opt.isDatetime {
			key += ":datetime"
		}
		fmt.Println(key)
		n++
	}
}

func (enc *headerEncoder) add(v reflect.Value, column, row int) {
	enc.cells = append(enc.cells, headerCell{})
}
