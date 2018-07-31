package timeproxy

import (
	"time"
)

type (
	Duration   = time.Duration
	Location   = time.Location
	Month      = time.Month
	ParseError = time.ParseError
	Time       = time.Time
	Weekday    = time.Weekday
)

const (
	ANSIC       = "Mon Jan _2 15:04:05 2006"
	UnixDate    = "Mon Jan _2 15:04:05 MST 2006"
	RubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
	RFC822      = "02 Jan 06 15:04 MST"
	RFC822Z     = "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
	RFC850      = "Monday, 02-Jan-06 15:04:05 MST"
	RFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"
	RFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
	RFC3339     = "2006-01-02T15:04:05Z07:00"
	RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
	Kitchen     = "3:04PM"
	Stamp       = "Jan _2 15:04:05"
	StampMilli  = "Jan _2 15:04:05.000"
	StampMicro  = "Jan _2 15:04:05.000000"
	StampNano   = "Jan _2 15:04:05.000000000"
)

const (
	Nanosecond  Duration = 1
	Microsecond          = 1000 * Nanosecond
	Millisecond          = 1000 * Microsecond
	Second               = 1000 * Millisecond
	Minute               = 60 * Second
	Hour                 = 60 * Minute
)

const (
	January Month = 1 + iota
	February
	March
	April
	May
	June
	July
	August
	September
	October
	November
	December
)

const (
	Sunday Weekday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

var (
	Local *Location = time.Local
	UTC   *Location = time.UTC
)

// TimeProxy is an interface that maps 1:1 with the standalone functions in
// this pacakge. By default, TimeProxy uses the functions from the time package
// in the standard library.
//
// Replace TimeProxy with a different implementation to gain control over time
// in tests.
var TimeProxy Proxy = RealTime{}

// Proxy is a proxy for the time package. By default, the methods from
// stdlib are called.
type Proxy interface {
	Now() Time
	Tick(Duration) <-chan Time
	After(Duration) <-chan Time
	Sleep(Duration)
	NewTicker(Duration) *Ticker
	AfterFunc(Duration, func()) *Timer
	NewTimer(Duration) *Timer
	Since(Time) Duration
	Until(Time) Duration
}

// Now calls TimeProxy.Now
func Now() Time {
	return TimeProxy.Now()
}

// Tick calls TimeProxy.Tick
func Tick(d Duration) <-chan Time {
	return TimeProxy.Tick(d)
}

// After calls TimeProxy.After
func After(d Duration) <-chan Time {
	return TimeProxy.After(d)
}

// Sleep calls TimeProxy.Sleep
func Sleep(d Duration) {
	TimeProxy.Sleep(d)
}

// NewTicker calls TimeProxy.NewTicker
func NewTicker(d Duration) *Ticker {
	return TimeProxy.NewTicker(d)
}

// RealTime dispatches all calls to the stdlib time package
type RealTime struct{}

// Now calls time.Now
func (RealTime) Now() Time {
	return time.Now()
}

// Sleep calls time.Sleep
func (RealTime) Sleep(d Duration) {
	time.Sleep(d)
}

// After calls time.After
func (RealTime) After(d Duration) <-chan Time {
	return time.After(d)
}

// Tick calls time.Tick
func (RealTime) Tick(d Duration) <-chan Time {
	//lint:ignore SA1015 (don't warn about leaking ticker)
	return time.Tick(d)
}

// NewTicker calls time.NewTicker
func (RealTime) NewTicker(d Duration) *Ticker {
	ticker := time.NewTicker(d)
	return &Ticker{
		C:        ticker.C,
		StopFunc: ticker.Stop,
	}
}

// AfterFunc calls time.AfterFunc
func (RealTime) AfterFunc(d Duration, f func()) *Timer {
	timer := time.AfterFunc(d, f)
	return &Timer{
		C:         timer.C,
		ResetFunc: timer.Reset,
		StopFunc:  timer.Stop,
	}
}

// NewTimer calls time.NewTimer. It returns a timeproxy Timer with StopFunc
// set to the timer's Stop method and ResetFunc set to the timer's Reset
// method.
func (RealTime) NewTimer(d Duration) *Timer {
	timer := time.NewTimer(d)
	return &Timer{
		C:         timer.C,
		ResetFunc: timer.Reset,
		StopFunc:  timer.Stop,
	}
}

// Since calls time.Since
func (RealTime) Since(t Time) Duration {
	return time.Since(t)
}

// Until calls time.Until
func (RealTime) Until(t Time) Duration {
	return time.Until(t)
}

// ParseDuration calls time.ParseDuration
func ParseDuration(s string) (Duration, error) {
	return time.ParseDuration(s)
}

// FixedZone calls time.FixedZone
func FixedZone(name string, offset int) *Location {
	return time.FixedZone(name, offset)
}

// LoadLocation calls time.LoadLocation
func LoadLocation(name string) (*Location, error) {
	return time.LoadLocation(name)
}

// LoadLocationFromTZData calls time LoadLocationFromTZData
func LoadLocationFromTZData(name string, data []byte) (*Location, error) {
	return time.LoadLocationFromTZData(name, data)
}

// Parse calls time.Parse
func Parse(layout, value string) (Time, error) {
	return time.Parse(layout, value)
}

// ParseInLocation calls time.ParseInLocation
func ParseInLocation(layout, value string, loc *Location) (Time, error) {
	return time.ParseInLocation(layout, value, loc)
}

// Date calls time.Date
func Date(year int, month Month, day, hour, min, sec, nsec int, loc *Location) Time {
	return time.Date(year, month, day, hour, min, sec, nsec, loc)
}

// Unix calls time.Unix
func Unix(sec, nsec int64) Time {
	return time.Unix(sec, nsec)
}

// Ticker is an abstraction of time.Ticker that can be patched more easily.
type Ticker struct {
	C        <-chan Time
	StopFunc func()
}

func (t *Ticker) Stop() {
	t.StopFunc()
}

// Timer is an abstraction of time.Timer that can be patched more easily.
type Timer struct {
	C         <-chan Time
	ResetFunc func(Duration) bool
	StopFunc  func() bool
}

// Reset calls t.ResetFunc
func (t *Timer) Reset(d Duration) bool {
	return t.ResetFunc(d)
}

// Stop calls t.StopFunc
func (t *Timer) Stop() bool {
	return t.StopFunc()
}

// Since calls TimeProxy.Since
func Since(t Time) Duration {
	return TimeProxy.Since(t)
}

// Until calls TimeProxy.Until
func Until(t Time) Duration {
	return TimeProxy.Until(t)
}

// NewTimer returns TimeProxy.NewTimer
func NewTimer(d Duration) *Timer {
	return TimeProxy.NewTimer(d)
}
