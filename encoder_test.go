package sheet

import (
	"fmt"
	"github.com/k0kubun/pp"
	"testing"
	"time"
)

type SampleMarshal struct {
	ID        string
	Num       int
	PID       *string
	Time      time.Time `sheet:"datetime"`
	List      []string
	UInt      uint
	Item      *SampleItem
	Bool      bool
	Hoges     []SampleHoge
	Hoge      SampleHoge
	Float     float32
	PList     []*string
	CreatedAt int64 `sheet:"datetime"`
}

type SampleItem struct {
	Code    string
	GraphID int
	Hoge    SampleHoge
}

type SampleHoge struct {
	Title string
	Order int
}

type SampleSlicePtr struct {
	List []*string
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
		//Item: &SampleItem{
		//	Code:    "code_01",
		//	GraphID: 1000,
		//	Hoge: SampleHoge{
		//		Title: "title_01",
		//		Order: 1,
		//	},
		//},
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
	}
	values, err := NewEncoder().Encode(sample)
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
		Float: 3.1415,
		PList: []*string{&pA, &pB},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewEncoder().Encode(sample)
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
	}
}

// 100000	     14261 ns/op	    5120 B/op	     125 allocs/op
// 200000	      9532 ns/op	    4288 B/op	      99 allocs/op
// 200000	      8230 ns/op	    2272 B/op	      93 allocs/op
