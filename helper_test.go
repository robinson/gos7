package gos7

import (
	"testing"
	"time"
)

func TestHelper_SetBoolAt(t *testing.T) {
	var h Helper
	input := []struct {
		in   byte
		out  byte
		pos  uint
		data bool
	}{
		{0b101, 0b111, 1, true},
		{0b111, 0b101, 1, false},
		{0b101, 0b001, 2, false},
		{0b111, 0b011, 2, false},
		{0b11111111, 0b11011111, 5, false},
	}

	for _, i := range input {
		b := h.SetBoolAt(i.in, i.pos, i.data)
		if b != i.out {
			t.Errorf("expected %b given %b", i.out, b)
		}
	}
}

func TestHelper_GetBoolAt(t *testing.T) {
	var h Helper
	input := []struct {
		in  byte
		pos uint
		out bool
	}{
		{0b101, 1, false},
		{0b111, 1, true},
		{0b101, 2, true},
		{0b011, 2, false},
		{0b11111111, 5, true},
	}

	for _, i := range input {
		b := h.GetBoolAt(i.in, i.pos)
		if b != i.out {
			t.Errorf("expected %v given %v", i.out, b)
		}
	}
}

// S7 counts the day of week from 1 (Sunday) to 7 (Saturday) in DATE_AND_TIME
// and DTL. 2026-11-15 is a Sunday.
var weekdayCases = []struct {
	day  int
	want byte
}{
	{15, 1}, {16, 2}, {17, 3}, {18, 4}, {19, 5}, {20, 6}, {21, 7},
}

func TestHelper_SetDateTimeAtWeekday(t *testing.T) {
	var h Helper
	for _, c := range weekdayCases {
		value := time.Date(2026, time.November, c.day, 13, 37, 42, 456000000, time.UTC)
		buffer := make([]byte, 8)
		h.SetDateTimeAt(buffer, 0, value)
		if got := buffer[7] & 0x0f; got != c.want {
			t.Errorf("%s: day of week %d, want %d", value.Weekday(), got, c.want)
		}
		if got := buffer[7] >> 4; got != 6 { // last digit of the 456 ms
			t.Errorf("%s: millisecond digit %d, want 6", value.Weekday(), got)
		}
	}
}

func TestHelper_SetDTLAtWeekday(t *testing.T) {
	var h Helper
	for _, c := range weekdayCases {
		value := time.Date(2026, time.November, c.day, 13, 37, 42, 0, time.UTC)
		buffer := make([]byte, 12)
		h.SetDTLAt(buffer, 0, value)
		if got := buffer[4]; got != c.want {
			t.Errorf("%s: day of week %d, want %d", value.Weekday(), got, c.want)
		}
	}
}
