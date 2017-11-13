package sheet

import (
	"reflect"
	"time"
)

const (
	tagName = "sheet"
)

var (
	typeOfTime = reflect.TypeOf(time.Time{})
)

func Marshal(v interface{}) ([][]interface{}, error) {
	return newEncoder().Encode(v)
}

func Unmarshal(formats [][]string, values [][]string, v interface{}) error {
	return newDecoder(formats).Decode(values, v)
}

func Header(v interface{}) {
	newHeaderEncoder().Encode(v)
}
