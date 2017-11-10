package sheet

import (
	"testing"
)

type Integer struct {
	Int   int   `sheet:"title=int"`
	Int8  int8  `sheet:"title=int8"`
	Int16 int16 `sheet:"title=int16"`
	Int32 int32 `sheet:"title=int32"`
	Int64 int64 `sheet:"title=int64"`
}

type UnsignedInteger struct {
	Uint   uint   `sheet:"title=uint"`
	Uint8  uint8  `sheet:"title=uint8"`
	Uint16 uint16 `sheet:"title=uint16"`
	Uint32 uint32 `sheet:"title=uint32"`
	Uint64 uint64 `sheet:"title=uint64"`
}

type Float struct {
	Float32 float32 `sheet:"title=float32"`
	Float64 float64 `sheet:"title=float64"`
}

type String struct {
	String string `sheet:"title=string"`
}

type Bool struct {
	True  bool `sheet:"title=true"`
	False bool `sheet:"title=false"`
}

type Slice struct {
	IDs  []int    `sheet:"title=lds"`
	List []string `sheet:"title=list"`
}

func TestMarshal(t *testing.T) {

}

func Test_Unmarshal(t *testing.T) {

}
