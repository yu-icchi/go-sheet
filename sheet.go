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
	return NewEncoder().Encode(v)
}

func Unmarshal(formats [][]string, values [][]string, v interface{}) error {
	return NewDecoder(formats).Decode(values, v)
}

func Header(v interface{}) {
	NewHeaderEncoder().Encode(v)
}
