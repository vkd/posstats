package main

import (
	"testing"
)

func TestMemoryStatsCounter_GetAll(t *testing.T) {
	ms := NewMemoryStatsCounter()
	testStats(t, ms)
}
