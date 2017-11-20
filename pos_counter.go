package main

import (
	"time"
)

// MapPosCount - in-memory pos counter
type MapPosCount struct {
	SeriesTimeout time.Duration
	ResetTimeout  time.Duration

	m map[string]*mapPosValue
}

var _ PosCounter = (*MapPosCount)(nil)

// NewMapPosCount - create in-memory pos counter
func NewMapPosCount(seriesTimeout, resetTimeout time.Duration) *MapPosCount {
	m := &MapPosCount{
		SeriesTimeout: seriesTimeout,
		ResetTimeout:  resetTimeout,

		m: make(map[string]*mapPosValue),
	}
	return m
}

// Pos - calc next pos
func (m *MapPosCount) Pos(ifa string) (int, error) {
	now := time.Now()

	p, ok := m.m[ifa]
	if !ok {
		m.m[ifa] = &mapPosValue{now, 1}
		return 1, nil
	}
	dt := now.Sub(p.last)
	p.last = now

	if dt > m.ResetTimeout {
		p.pos = 1
		return 1, nil
	}
	if dt > m.SeriesTimeout {
		p.pos++
	}
	return p.pos, nil
}

type mapPosValue struct {
	last time.Time
	pos  int
}
