package merge

import "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/profile"

type Profile interface {
	GetCounters(pkgFile string) Counters // return nil to indicate not-existent
	RangeCounters(func(pkgFile string, counters Counters) bool)

	SetCounters(pkgFile string, counters Counters)
	Clone() Profile
}
type Counters interface {
	Len() int
	Range(func(i int, counter Counter) bool)

	Get(i int) Counter
	Set(i int, counter Counter)
	New(n int) Counters
}

// a Counter represents a block
type Counter interface {
	// Add called when two blocks are the same
	Add(b Counter) Counter
}
type IntCounter int
type LabelCounter map[string]int

func (c IntCounter) Add(b Counter) Counter { return IntCounter(c + b.(IntCounter)) }
func (c LabelCounter) Add(b Counter) Counter {
	r := make(LabelCounter, len(c))
	for k, v := range c {
		r[k] = v
	}
	bm := b.(LabelCounter)
	for k, v := range bm {
		r[k] += v
	}
	return r
}

type StdProfile struct {
	profile  *profile.Profile
	counters map[string]Counters
}

func NewStdProfile(profile *profile.Profile) Profile {
	return &StdProfile{profile: profile, counters: NewIntCountersMapping(profile.Counters())}
}
func NewIntCountersMapping(counters map[string][]int) map[string]Counters {
	m := make(map[string]Counters, len(counters))
	for k, v := range counters {
		m[k] = NewIntCounters(v)
	}
	return m
}

// Clone implements Profile
func (c *StdProfile) Clone() Profile {
	return NewStdProfile(c.profile.Clone())
}

// GetCounters implements Profile
func (c *StdProfile) GetCounters(pkgFile string) Counters {
	counters := c.counters[pkgFile]
	if counters == nil {
		return nil
	}
	return counters
}

// RangeCounters implements Profile
func (c *StdProfile) RangeCounters(fn func(pkgFile string, counters Counters) bool) {
	for pkgFile, counters := range c.counters {
		if !fn(pkgFile, counters) {
			return
		}
	}
}

// SetCounters implements Profile
func (c *StdProfile) SetCounters(pkgFile string, counters Counters) {
	c.counters[pkgFile] = counters
}

func (c *StdProfile) Output() *profile.Profile {
	newCounters := make(map[string][]int, len(c.counters))
	for pkgFile, counters := range c.counters {
		newCounters[pkgFile] = counters.(IntCounters).ToInts()
	}
	c.profile.ResetCounters(newCounters)
	return c.profile
}

type IntCounters []IntCounter

func NewIntCounters(counters []int) IntCounters {
	c := make(IntCounters, 0, len(counters))
	for _, v := range counters {
		c = append(c, IntCounter(v))
	}
	return c
}

// Get implements Counters
func (c IntCounters) Get(i int) Counter {
	return c[i]
}

// Len implements Counters
func (c IntCounters) Len() int {
	return len(c)
}

// New implements Counters
func (c IntCounters) New(n int) Counters {
	return make(IntCounters, n)
}

// Range implements Counters
func (c IntCounters) Range(fn func(i int, counter Counter) bool) {
	for i, counter := range c {
		if !fn(i, counter) {
			return
		}
	}
}

// Set implements Counters
func (c IntCounters) Set(i int, counter Counter) {
	c[i] = counter.(IntCounter)
}

func (c IntCounters) ToInts() []int {
	ints := make([]int, 0, len(c))
	for _, i := range c {
		ints = append(ints, int(i))
	}
	return ints
}
