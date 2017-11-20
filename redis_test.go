package main

import (
	"testing"
	"time"
)

func TestRedisClient_Pos(t *testing.T) {
	t.Skip()
	seriesMs := 1000
	resetMs := 2000
	pc, err := NewRedisClient(":6379", time.Duration(seriesMs)*time.Millisecond, time.Duration(resetMs)*time.Millisecond)
	if err != nil {
		t.Fatalf("Error on create redis client: %v", err)
	}
	pc.flushAll()

	_ = pc
	testPosCounterPos(t, pc, seriesMs, resetMs)
}

func TestRedisClient_GetAll(t *testing.T) {
	t.Skip()
	rc, err := NewRedisClient(":6379", time.Second, 2*time.Second)
	if err != nil {
		t.Fatalf("Error on create redis client: %v", err)
	}
	rc.flushAll()

	testStats(t, rc)
}

type testStatser interface {
	StatsGetter
	StatsCollector
}

func testStats(t *testing.T, rc testStatser) {
	var err error
	for _, tt := range []struct {
		app, platform, country string
	}{
		{"12345", "ios", "USA"},
		{"12345", "ios", "USA"},
		{"12345", "android", "RUS"},
		{"abcdef", "ios", "RUS"},
		{"abcdef", "ios", "RUS"},
		{"abcdef", "ios", "RUS"},
	} {
		err = rc.Add(tt.app, tt.platform, tt.country)
		if err != nil {
			t.Errorf("Added stat: %s, %s, %s", tt.app, tt.platform, tt.country)
			t.Fatalf("Error on add stat: %v", err)
		}
	}

	stats, err := rc.GetAll()
	if err != nil {
		t.Fatalf("Error on get stats: %v", err)
	}

	ss := []*Stat{
		{"12345", "ios", "USA", 2},
		{"12345", "android", "RUS", 1},
		{"abcdef", "ios", "RUS", 3},
	}

	if len(stats) != len(ss) {
		t.Errorf("Wrong stats len: %v", stats)
	}

	for i, s := range ss {
		if stats[i].App != s.App {
			t.Errorf("[%d] Wrong app: %s", i, stats[i].App)
		}
		if stats[i].Platform != s.Platform {
			t.Errorf("[%d] Wrong platform: %s", i, stats[i].Platform)
		}
		if stats[i].Country != s.Country {
			t.Errorf("[%d] Wrong country: %s", i, stats[i].Country)
		}
		if stats[i].Count != s.Count {
			t.Errorf("[%d] Wrong count: %d", i, stats[i].Count)
		}
	}
}
