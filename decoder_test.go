package sheet

import (
	"fmt"
	"github.com/k0kubun/pp"
	"testing"
	"time"
)

type SampleUnmarshal struct {
	ID    string                `sheet:"id,index"`
	Sub   *SampleUnmarshalSub   `sheet:"sub"`
	Num   int                   `sheet:"num"`
	Arr   []string              `sheet:"arr,csv"`
	PID   *string               `sheet:"pid"`
	List  []*string             `sheet:"list"`
	SList []SampleUnmarshalSub2 `sheet:"slist"`
	Now   time.Time             `sheet:"now,datetime"`
}

type SampleUnmarshalSub struct {
	Code string `sheet:"code"`
	Num  int    `sheet:"num"`
}

type SampleUnmarshalSub2 struct {
	Code string `sheet:"code"`
	Num  int    `sheet:"num"`
}

func TestNewDecoder(t *testing.T) {
	formats := [][]string{
		{"id", "sub", "", "num", "arr:csv", "pid", "list", "slist", "", "", "now:datetime"},
		{"", "code", "num", "", "", "", "", "_index", "code", "num"},
	}
	values := [][]string{
		{"id_01", "aaa", "123456789", "123", "AA,BB,CC", "p-id", "AA", "1", "", "90", "2017-11-06 01:27:00"},
		{"", "", "", "", "", "", "BB", "2", "code_1_02", "12"},
		{"", "", "", "", "", "", "CC", "3", "code_1_03", "13"},
	}
	sample := &SampleUnmarshal{}
	err := newDecoder(formats).Decode(values, sample)
	fmt.Println(err)
	pp.Println(sample)
}

func BenchmarkNewDecoder(b *testing.B) {
	formats := [][]string{
		{"id", "sub", "", "num", "arr:csv", "pid", "list", "slist", "", "", "now:datetime"},
		{"", "code", "num", "", "", "", "", "_index", "code", "num"},
	}
	values := [][]string{
		{"id_01", "code_01", "1100", "1", "AA,BB,CC", "p_id_01", "AA", "1", "", "", "2017-11-06 01:27:00"},
		{"", "", "", "", "", "", "BB", "2", "code_1_02", ""},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sample := &SampleUnmarshal{}
		newDecoder(formats).Decode(values, sample)
	}
}

// 200000	      7158 ns/op	    1312 B/op	      52 allocs/op
// 100000	     13480 ns/op	    3249 B/op	      95 allocs/op
