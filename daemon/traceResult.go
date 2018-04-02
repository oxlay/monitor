package agent

import (
	"time"
)

// TraceResults represents all the trace results for a given website.
// The retention policy of those results is user-defined:
// the RetainedResults config parameter specifies how many TraceResults to keep.
type TraceResults []TraceResult

// A TraceResult represents the results of a request to a website.
// It contains timing information about the different phases of the request,
// as well as the request result (error or HTTP response code).
type TraceResult struct {
	// Date is the date at which the request's response was received.
	Date time.Time

	// DNSTime is the duration of the DNS lookup.
	DNSTime time.Duration

	// TLSTime is the duration of the TLS handshake, if applicable.
	// If the website was contacted over HTTP, TLStime will be set to time.Duration(0).
	TLSTime time.Duration

	// ConnectTime is the TCP connection time.
	// TODO: check this
	ConnectTime time.Duration

	// TTFB is the time to first byte.
	// It is computed from the start of the request and therefore includes other
	// durations such as DNS lookup time.
	TTFB time.Duration

	// Error stores the error if the request resulted in an error, or nil otherwise.
	Error error

	// StatusCode stores the HTTP response code of the request, or 0 if the request
	// resulted in a (non-HTTP) error.
	StatusCode int
}

// StartIndexFor(timespan) returns the index (startIndex) of the first
// trace result that is included in the provided timespan.
// In other words, t[startIndex:] will be the metrics obtained between [now, now - timespan].
//
// It leverages the fact that TraceResults are sorted by increasing date.
// The returned startIdx can then be used to aggregate the metrics fetched
// during the specified timespan.
//
// For example, given the following TestResults:
//		[]TraceResult{
//			{ currentTime - 6 minutes, ... }
//			{ currentTime - 4 minutes, ... }
//			{ currentTime - 2 minutes, ... }
//			{ currentTime, ... },
//		}
// and given timespan = 180 (seconds), StartIndexFor(Timespan) would return 2,
// as it is the index of the first TraceResult of the slice
// that occured in the timeframe [now, now - 180 seconds]
func (t TraceResults) StartIndexFor(timespan int) int {
	threshold := time.Now().Add(-time.Duration(timespan) * time.Second)
	for i := len(t) - 1; i >= 0; i-- {
		if t[i].Date.Before(threshold) {
			return i + 1 // TODO: handle case where i + 1 is out of range
		}
	}
	return 0
}

// TTFBs extracts the TTFB of each of the trace results, starting from startIdx.
// It returns those TTFB values in a slice.
func (t TraceResults) TTFBs(startIdx int) (durations []time.Duration) {
	for i := startIdx; i < len(t); i++ {
		durations = append(durations, t[i].TTFB)
	}
	return
}

// CountCodes counts the HTTP response codes in the latest trace results, starting from startIdx.
// The return value maps from each HTTP response code encountered to the number of such codes.
func (t TraceResults) CountCodes(startIdx int) map[int]int {
	codesCount := make(map[int]int)
	for i := startIdx; i < len(t); i++ {
		code := t[i].StatusCode
		if code != 0 {
			// If the request led to an HTTP response code, and not an error
			codesCount[code]++
		}
	}
	return codesCount
}

// CountErrors counts the errors in the latest trace results, starting from startIdx.
// The return value maps from each error string encountered to the number of such errors.
func (t TraceResults) CountErrors(startIdx int) map[string]int {
	errorsCount := make(map[string]int)
	for i := startIdx; i < len(t); i++ {
		error := t[i].Error
		if error != nil {
			errorsCount[error.Error()]++
		}
	}
	return errorsCount
}

// Availability returns the availability based on the latest trace results, starting from startIdx.
// The return value is between 0 and 1.
func (t TraceResults) Availability(startIdx int) float64 {
	c := 0
	for i := startIdx; i < len(t); i++ {
		if t[i].IsValid() {
			c++
		}
	}
	return float64(c) / float64(len(t)-startIdx)
}

// IsValid returns whether the trace result is considered valid or not.
//
// To be considered valid, the associated request must satisfy two conditions:
// the request did not end with an error, and
// the HTTP response is neither a Client error nor a Server error.
func (t *TraceResult) IsValid() bool {
	return (t.Error == nil) && (t.StatusCode < 400)
}