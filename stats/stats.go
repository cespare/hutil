package stats

import (
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	discreteBucketSizeSeconds = 10
	recordBufSize             = 16
	numDiscreteBuckets        = (60 / discreteBucketSizeSeconds) * 60 * 24 // matches the biggest time period
)

var (
	timePeriods = []timePeriod{
		{"last ten seconds", 1},
		{"last five minutes", (60 / discreteBucketSizeSeconds) * 5},
		{"last hour", (60 / discreteBucketSizeSeconds) * 60},
		{"last day", (60 / discreteBucketSizeSeconds) * 60 * 24},
	}
)

type timePeriod struct {
	Name    string
	Buckets int
}

type discretePoint struct {
	value string
	count uint64
}

// We're going to use a slice instead of a map here because we expect the number of values to be small. We
// could instead use a map[string]uint64 for discretePoints, but that's probably not going to be particularly
// cpu- or memory-efficient when we have many thousands of them.
type discretePoints []discretePoint

func (ps *discretePoints) inc(value string) {
	for i, p := range *ps {
		if p.value == value {
			(*ps)[i].count++
			return
		}
	}
	*ps = append(*ps, discretePoint{value, 1})
}

// 10s of discrete values.
type discreteValuesBucket struct {
	total          uint64
	responseStatus *discretePoints
}

func (b *discreteValuesBucket) clear() {
	d := discretePoints([]discretePoint{})
	*b = discreteValuesBucket{
		responseStatus: &d,
		total:          0,
	}
}

type discreteValuesBuffer struct {
	ring    [numDiscreteBuckets]discreteValuesBucket
	current int
	full    bool
}

func (b *discreteValuesBuffer) currentBucket() *discreteValuesBucket {
	return &b.ring[b.current]
}

func (b *discreteValuesBuffer) lastNBuckets(n int) <-chan *discreteValuesBucket {
	available := 0
	if b.full {
		available = numDiscreteBuckets
	} else {
		available = b.current + 1
	}
	if n > available {
		n = available
	}
	c := make(chan *discreteValuesBucket)
	go func() {
		for i := n - 1; i >= 0; i-- {
			idx := (b.current + numDiscreteBuckets - i) % numDiscreteBuckets
			c <- &b.ring[idx]
		}
		close(c)
	}()
	return c
}

// Rotate advances the ring buffer to the next bucket and zeroes it out.
func (b *discreteValuesBuffer) rotate() {
	b.current = (b.current + 1) % numDiscreteBuckets
	if !b.full && b.current == 0 {
		b.full = true
	}
	b.currentBucket().clear()
}

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
	http.Handler

	records  chan *statRecord
	discrete *discreteValuesBuffer
}

type DiscreteCount struct {
	Total          []uint64
	ResponseStatus map[string][]uint64
}

func newDiscreteCount() *DiscreteCount {
	return &DiscreteCount{
		ResponseStatus: make(map[string][]uint64),
		Total: make([]uint64, len(timePeriods)),
	}
}

type Summary struct {
	*DiscreteCount
}

func New(handler http.Handler) *Stats {
	buf := &discreteValuesBuffer{}
	buf.currentBucket().clear()
	s := &Stats{
		Handler:  handler,
		records:  make(chan *statRecord, recordBufSize),
		discrete: buf,
	}
	go s.process()
	return s
}

func (s *Stats) process() {
	t := time.NewTicker(discreteBucketSizeSeconds * time.Second)
	for {
		select {
		case <-t.C:
			s.discrete.rotate()
		case r := <-s.records:
			s.discrete.currentBucket().total++
			status := strconv.Itoa(r.status)
			s.discrete.currentBucket().responseStatus.inc(status)
		}
	}
}

func (s *Stats) sumDiscrete(count *DiscreteCount, timePeriodIdx int) {
	nBuckets := timePeriods[timePeriodIdx].Buckets
	for b := range s.discrete.lastNBuckets(nBuckets) {
		if b.responseStatus == nil {
			continue
		}
		count.Total[timePeriodIdx] += (*b).total
		for _, p := range *b.responseStatus {
			statuses := count.ResponseStatus[p.value]
			if statuses == nil {
				count.ResponseStatus[p.value] = make([]uint64, len(timePeriods))
			}
			count.ResponseStatus[p.value][timePeriodIdx] += p.count
		}
	}
}

func (s *Stats) summary() *Summary {
	count := newDiscreteCount()
	for i := range timePeriods {
		s.sumDiscrete(count, i)
	}
	return &Summary{count}
}

func (s *Stats) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	record := &statRecord{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
	start := time.Now()
	s.Handler.ServeHTTP(record, r)
	record.elapsed = time.Since(start)
	s.records <- record
}

func (s *Stats) HandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		context := struct {
			TimePeriods []timePeriod
			Summary *Summary
		}{
			timePeriods,
			s.summary(),
		}
		if err := page.Execute(w, context); err != nil {
			log.Println(err)
			http.Error(w, "Error writing server stats.", http.StatusInternalServerError)
			return
		}
	}
}
