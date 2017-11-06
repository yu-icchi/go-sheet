package sheet

import (
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
		{"ID", "Sub", "", "Num", "PID", "List", "SList", "", "", "Now"},
		{"", "Code", "Num", "", "", "", "_index", "Code", "Num"},
	}
	values := [][]string{
		{"id_01", "code_01", "1100", "1", "p_id_01", "AA", "1", "", "", "2017-11-06 01:27:00"},
		{"", "", "", "", "", "BB", "2", "code_1_02", ""},
	}
	sample := &SampleUnmarshal{}
	NewDecoder(formats).Decode(values, sample)
	pp.Println(sample)
}
