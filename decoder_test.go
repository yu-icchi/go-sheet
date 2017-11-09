package sheet

import (
	"fmt"
	"github.com/k0kubun/pp"
	"testing"
	"time"
)

type SampleUnmarshal struct {
	ID    string
	Sub   SampleUnmarshalSub
	Num   int
	PID   *string
	List  []*string
	SList []SampleUnmarshalSub
	Now   time.Time
}

type SampleUnmarshalSub struct {
	Code string
	Num  int
}

func TestNewDecoder(t *testing.T) {
	formats := [][]string{
		{"ID", "Sub", "", "Num", "PID", "List", "SList", "", "", "Now:datetime"},
		{"", "Code", "Num", "", "", "", "_index", "Code", "Num"},
	}
	values := [][]string{
		{"id_01", "aaa", "123456789", "123", "", "AA", "1", "", "", "2017-11-06 01:27:00"},
		{"", "", "", "", "", "", "2", "code_1_02", "12"},
		{"", "", "", "", "", "CC", "3", "code_1_03", "13"},
	}
	sample := &SampleUnmarshal{}
	err := NewDecoder(formats).Decode(values, sample)
	fmt.Println(err)
	pp.Println(sample)
}

func BenchmarkNewDecoder(b *testing.B) {
	formats := [][]string{
		{"ID", "Sub", "", "Num", "PID", "List", "SList", "", "", "Now:datetime"},
		{"", "Code", "Num", "", "", "", "_index", "Code", "Num"},
	}
	values := [][]string{
		{"id_01", "code_01", "1100", "1", "p_id_01", "AA", "1", "", "", "2017-11-06 01:27:00"},
		{"", "", "", "", "", "BB", "2", "code_1_02", ""},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sample := &SampleUnmarshal{}
		NewDecoder(formats).Decode(values, sample)
	}
}

// 200000	      6285 ns/op	    1800 B/op	      62 allocs/op
// 200000	      5350 ns/op	    1560 B/op	      49 allocs/op
// 300000	      4733 ns/op	    1256 B/op	      41 allocs/op
// 300000	      4639 ns/op	    1176 B/op	      40 allocs/op
// 300000	      4735 ns/op	    1208 B/op	      39 allocs/op

// refactor
// 200000	      5939 ns/op	    1328 B/op	      52 allocs/op
