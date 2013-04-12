package hutil

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// TODO: possibly interesting:
// * response status
// * elapsed time
// * request path (need to normalize, which is hard)
// * response size

type statRecord struct {
	http.ResponseWriter

	elapsed time.Duration
	status  int
}

func (r *statRecord) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

type Stats struct {
	count uint64
	continuousValues map[string]uint64
	discreteValues map[string]map[string]uint64
}

func (s *Stats) Copy() *Stats {
	return &Stats{
		count: s.count,
	}
}

func (s *Stats) merge(r *statRecord) {
	s.count++
	s.continuousValues["elapsedMillis"] += uint64(r.elapsed.Nanoseconds() / 1000 / 1000)
	status := strconv.Itoa(r.status)
	statuses, ok := s.discreteValues["statuses"]
	if !ok {
		s.discreteValues["statuses"] = map[string]uint64{status: uint64(1)}
	} else {
		statuses[status] += uint64(1)
	}
}

type StatRecorder struct {
	http.Handler
	*sync.RWMutex

	stats *Stats
}

func NewStatRecorder(h http.Handler) *StatRecorder {
	return &StatRecorder{
		Handler: h,
		RWMutex: &sync.RWMutex{},
		stats: &Stats{
			count: 0,
			continuousValues: make(map[string]uint64),
			discreteValues: make(map[string]map[string]uint64),
		},
	}
}

func (s *StatRecorder) record(r *statRecord) {
	s.Lock()
	defer s.Unlock()
	s.stats.merge(r)
}

func (s *StatRecorder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	record := &statRecord{
		ResponseWriter: w,
		status:         http.StatusOK,
		elapsed:        time.Duration(0),
	}

	startTime := time.Now()
	s.Handler.ServeHTTP(record, r)
	record.elapsed = time.Since(startTime)

	s.record(record)
}

func (s *StatRecorder) Snapshot() *Stats {
	s.RLock()
	defer s.RUnlock()
	return s.stats.Copy()
}

func (s *StatRecorder) HandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := s.Snapshot()
		fmt.Fprintf(w, "Count: %d\n", stats.count)
		mean := float64(stats.continuousValues["elapsedMillis"]) / float64(stats.count)
		fmt.Fprintf(w, "Mean elapsed milliseconds: %5.3f\n", mean)
		fmt.Fprintln(w, "Status:")
		for name, value := range stats.discreteValues["statuses"] {
			fmt.Fprintf(w, "  %s %d\n", name, value)
		}
	}
}
