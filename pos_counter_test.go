package main

import (
	"testing"
	"time"
)

func TestMapPosCount_Pos(t *testing.T) {
	pc := NewMapPosCount(20*time.Millisecond, 100*time.Millisecond)
	testPosCounterPos(t, pc, 20, 100)
}

func testPosCounterPos(t *testing.T, pc PosCounter, seriesMs, resetMs int) {
	ifa := "test_ifa"
	for i, tt := range []struct {
		dtMs int
		pos  int
	}{
		0: {0, 1},
		1: {seriesMs / 2, 1},
		2: {seriesMs - 10, 1},
		3: {seriesMs + 10, 2},
		4: {resetMs - 10, 3},
		5: {resetMs + 10, 1},
	} {
		time.Sleep(time.Duration(tt.dtMs) * time.Millisecond)

		pos, err := pc.Pos(ifa)
		if err != nil {
			t.Fatalf("Error on pos ifa: %v", err)
		}
		if pos != tt.pos {
			t.Fatalf("[%d] Wrong pos: %d (want: %d)", i, pos, tt.pos)
		}
	}
}
