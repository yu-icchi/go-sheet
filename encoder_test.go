package sheet

import (
	"fmt"
	"testing"
	"time"

	"github.com/k0kubun/pp"
)

type SampleMarshal struct {
	ID        string
	Num       int
	PID       *string
	Time      time.Time `sheet:"datetime,title=time"`
	List      []string  `sheet:"csv"`
	UInt      uint
	Item      *SampleItem
	Bool      bool
	Hoges     []SampleHoge
	Hoge      SampleHoge
	Float     float32
	PList     []*string
	CreatedAt int64     `sheet:"datetime"`
	Floats    []float32 `sheet:"csv"`
}

type SampleItem struct {
	Code    string
	GraphID int
	Hoge    SampleHoge
	Slice   SampleSlicePtr
}

type SampleHoge struct {
	Title string
	Order int
}

type SampleSlicePtr struct {
	List []string `sheet:"csv"`
}

type SampleArrayPtr struct {
	ID   string
	List [2]SampleHoge
	Time time.Time
}

func TestNewEncoder(t *testing.T) {
	pid := "p_id"
	pA := "pa"
	pB := "pb"
	sample := &SampleMarshal{
		ID:   "test_marshal_2",
		Num:  100,
		PID:  &pid,
		Time: time.Now(),
		List: []string{"A", "B"},
		UInt: 90,
		Item: &SampleItem{
			Code:    "code_01",
			GraphID: 1000,
			Hoge: SampleHoge{
				Title: "title_01",
				Order: 1,
			},
			Slice: SampleSlicePtr{
				List: []string{"AA", "BB", "CC"},
			},
		},
		Bool: true,
		Hoges: []SampleHoge{
			{
				Title: "title_sub_1",
				Order: 1,
			},
			{
				Title: "title_sub_2",
				Order: 2,
			},
			{
				Title: "title_sub_3",
				Order: 3,
			},
		},
		Hoge: SampleHoge{
			Title: "title_02",
			Order: 2,
		},
		Float:     3.1415,
		PList:     []*string{&pA, &pB},
		CreatedAt: time.Now().Unix(),
		Floats:    []float32{1.1002, 2.21, 3.32, 5.67},
	}
	values, err := newEncoder().Encode(sample)
	fmt.Println(err)
	pp.Println(values)
}

func BenchmarkNewEncoder(b *testing.B) {
	pid := "p_id"
	pA := "pa"
	pB := "pb"
	sample := &SampleMarshal{
		ID:   "test_marshal_2",
		Num:  100,
		PID:  &pid,
		Time: time.Now(),
		List: []string{"A", "B"},
		UInt: 90,
		Item: &SampleItem{
			Code:    "code_01",
			GraphID: 1000,
			Hoge: SampleHoge{
				Title: "title_01",
				Order: 1,
			},
		},
		Bool: true,
		Hoges: []SampleHoge{
			{
				Title: "title_sub_1",
				Order: 1,
			},
			{
				Title: "title_sub_2",
				Order: 2,
			},
			{
				Title: "title_sub_3",
				Order: 3,
			},
		},
		Hoge: SampleHoge{
			Title: "title_02",
			Order: 2,
		},
		Float:  3.1415,
		PList:  []*string{&pA, &pB},
		Floats: []float32{1.1002, 2.21, 3.32, 5.67},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := newEncoder().Encode(sample)
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
	}
}

// 200000	      8682 ns/op	    2129 B/op	      86 allocs/op

// 200000	      7388 ns/op	    1648 B/op	      70 allocs/op
// 200000	      7567 ns/op	    1616 B/op	      69 allocs/op
