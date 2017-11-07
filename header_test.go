package sheet

import "testing"

type SampleHeader struct {
	ID        string
	UpdatedAt int64 `sheet:"datetime"`
}

func TestNewHeaderEncoder(t *testing.T) {
	sample := &SampleHeader{}
	NewHeaderEncoder().Encode(sample)
}
