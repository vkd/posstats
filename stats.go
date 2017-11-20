package main

// Stat - stat object
type Stat struct {
	App      string `json:"app"`
	Platform string `json:"platform"`
	Country  string `json:"country"`
	Count    int    `json:"count"`
}

// MemoryStatsCounter - in-memory stats counter
type MemoryStatsCounter struct {
	m  map[stat]*Stat
	ss []*Stat
}

type stat struct {
	App      string
	Platform string
	Country  string
}

// NewMemoryStatsCounter - create new in-memory stats counter
func NewMemoryStatsCounter() *MemoryStatsCounter {
	m := &MemoryStatsCounter{
		m: make(map[stat]*Stat),
	}
	return m
}

// Add - add stat
func (m *MemoryStatsCounter) Add(app, platform, country string) error {
	if s, ok := m.m[stat{app, platform, country}]; ok {
		s.Count++
		return nil
	}
	s := &Stat{App: app, Platform: platform, Country: country, Count: 1}
	m.m[stat{app, platform, country}] = s
	m.ss = append(m.ss, s)
	return nil
}

// GetAll - get all stats
func (m *MemoryStatsCounter) GetAll() ([]*Stat, error) {
	return m.ss, nil
}
