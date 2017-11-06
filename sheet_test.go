package sheet

import (
	"fmt"
	"github.com/k0kubun/pp"
	"testing"
)

type Sample struct {
	ID   int    `sheet:"id"`
	List []Item `sheet:"list"`
}

type Sample2 struct {
	ID      string `sheet:"id"`
	Num     int    `sheet:"num"`
	Sub     Sub2   `sheet:"sub"`
	Boo     bool   `sheet:"bool"`
	SubList []Sub2
}

type Sub2 struct {
	Code string `sheet:"code"`
	Age  int    `sheet:"age"`
}

type Term struct {
	Start int64 `sheet:"start"`
	End   int64 `sheet:"end"`
}

type STerm struct {
	Daykey string `sheet:"daykey"`
}

type ID struct {
	Num int    `sheet:"num"`
	Key string `sheet:"key"`
}

type Item struct {
	ID    string   `sheet:"itemId"`
	Count int      `sheet:"count"`
	IDs   []string `sheet:"ids"`
}

func TestMarshal(t *testing.T) {
	sample := &Sample{
		ID:   90,
		List: []Item{},
	}
	cells, err := Marshal(sample)
	fmt.Println(err)
	pp.Println(cells)
}

func Test_Unmarshal(t *testing.T) {
	formats := [][]string{
		{"ID", "List", "", ""},
		{"", "ID", "Count", "IDs"},
	}
	values := [][]string{
		{"90", "oo", "89", "A"},
		{"", "", "", "B"},
	}
	sample := &Sample{}
	Unmarshal(formats, values, sample)
	pp.Println(sample)
}
