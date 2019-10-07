package gos7

import (
	"testing"
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
