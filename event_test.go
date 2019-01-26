package observer

import (
	"testing"
	"time"
)

func TestDecTimeFormatting(t *testing.T) {
	floatTime := 1548515808.569242
	sec := int64(floatTime)
	nsec := int64(floatTime*float64(time.Millisecond)) % sec
	if sec != 1548515808 {
		t.Fatalf("expected 1548515808, got %d", sec)
	}
	if nsec != 569242 {
		t.Fatalf("expected 1548515808, got %d", nsec)
	}
}
