package httpcache

import (
	"iter"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ResponseInfo struct {
	// The time the request began.
	RequestTime time.Time

	// The time the response was received.
	ResponseTime time.Time

	// The date from the response's Date header.
	Date time.Time

	// The value from the Age header.
	Age time.Duration

	// Whether a max-age was set via the Cache-Control header.
	HasMaxAge bool

	// The max-age value from the Cache-Control header.
	MaxAge time.Duration

	// The value of the Expires header.
	Expires time.Time

	// The Last-Modified header value.
	LastModified time.Time

	// The ETag header value.
	ETag string
}

type lifetimeOption func(ri ResponseInfo, now time.Time, exp *time.Time)

// Set values from a [http.Response].
func (ri *ResponseInfo) UpdateFromResponse(resp *http.Response) {
	if date := getResponseDate(resp); !date.IsZero() {
		ri.Date = date
	}

	if age, hasAge := getResponseAge(resp); hasAge {
		ri.Age = age
	}

	if maxAge, hasMaxAge := getResponseMaxAge(resp); hasMaxAge {
		ri.MaxAge = maxAge
		ri.HasMaxAge = true
	}

	if expires := getResponseExpires(resp); !expires.IsZero() {
		ri.Expires = expires
	}

	if lastMod := getResponseLastModified(resp); !lastMod.IsZero() {
		ri.LastModified = lastMod
	}

	if etag := resp.Header.Get("ETag"); etag != "" {
		ri.ETag = etag
	}
}

// Get the expiration date.
func (ri ResponseInfo) GetExpiration(now time.Time, opts ...lifetimeOption) time.Time {
	curAge := ri.GetCurrentAge(now)
	lifetime := ri.GetFreshnessLifetime()
	remaining := lifetime - curAge
	exp := now.Add(remaining)

	for _, opt := range opts {
		opt(ri, now, &exp)
	}

	return exp
}

// Option to enforce a minimum cache duration.
func WithMinCacheDuration(min time.Duration) lifetimeOption {
	return func(ri ResponseInfo, now time.Time, exp *time.Time) {
		if exp.Sub(now) < min {
			*exp = now.Add(min)
		}
	}
}

// Option to enforce a maximum cache lifetime.
func WithMaxCacheDuration(max time.Duration) lifetimeOption {
	return func(ri ResponseInfo, now time.Time, exp *time.Time) {
		if exp.Sub(now) > max {
			*exp = now.Add(max)
		}
	}
}

// Get the freshness lifetime for the response.
func (ri ResponseInfo) GetFreshnessLifetime() time.Duration {
	if ri.HasMaxAge {
		return ri.MaxAge
	} else if !ri.Expires.IsZero() && (!ri.Date.IsZero() || !ri.ResponseTime.IsZero()) {
		date := ri.Date
		if date.IsZero() {
			date = ri.ResponseTime
		}

		return ri.Expires.Sub(date)
	} else {
		return 0
	}
}

// Get the current age of the request.
// See https://www.rfc-editor.org/info/rfc9111/#section-4.2.3
func (ri ResponseInfo) GetCurrentAge(now time.Time) time.Duration {
	responseTime := ri.ResponseTime
	if responseTime.IsZero() {
		responseTime = now
	}

	reqTime := ri.RequestTime
	if reqTime.IsZero() {
		reqTime = responseTime
	}

	dateVal := ri.Date
	if dateVal.IsZero() {
		dateVal = responseTime
	}

	apparentAge := max(0, responseTime.Sub(dateVal))

	responseDelay := responseTime.Sub(reqTime)
	correctedAge := ri.Age + responseDelay

	correctedInitialAge := max(apparentAge, correctedAge)

	residentTime := now.Sub(responseTime)

	return correctedInitialAge + residentTime
}

// Get the response date.
func getResponseDate(resp *http.Response) time.Time {
	dateStr := resp.Header.Get("Date")
	if dateStr != "" {
		respDate, err := http.ParseTime(dateStr)
		if err == nil {
			return respDate
		}
	}

	return time.Time{}
}

func getResponseAge(resp *http.Response) (time.Duration, bool) {
	ageStr := resp.Header.Get("Age")
	if ageStr != "" {
		ageInt, err := strconv.Atoi(ageStr)
		if err == nil {
			return time.Duration(ageInt) * time.Second, true
		}
	}

	return 0, false
}

// Get the Expires header date.
func getResponseExpires(resp *http.Response) time.Time {
	expHeader := resp.Header.Get("Expires")
	if expHeader != "" {
		parsed, err := http.ParseTime(expHeader)
		if err == nil {
			return parsed
		}
	}

	return time.Time{}
}

// Get the max-age from a request's Cache-Control header.
func getResponseMaxAge(resp *http.Response) (time.Duration, bool) {
	cc := resp.Header.Get("Cache-Control")
	if cc != "" {
		return parseMaxAge(parseCacheControl(cc))
	} else {
		return 0, false
	}
}

// Parse the max-age value from a sequence of Cache-Control key/values.
func parseMaxAge(kv iter.Seq2[string, string]) (maxAge time.Duration, ok bool) {
	for k, v := range kv {
		if k == "max-age" {
			maxAgeInt, err := strconv.Atoi(v)
			if err == nil {
				maxAge = time.Duration(maxAgeInt) * time.Second
				ok = true
				return
			}
		}
	}
	return
}

// Parse the Cache-Control header value into a sequence of key/value pairs.
func parseCacheControl(cc string) iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for part := range strings.SplitSeq(cc, ",") {
			part = strings.TrimSpace(part)
			k, v, found := strings.Cut(part, "=")
			if !found {
				continue
			}

			if !yield(k, v) {
				return
			}
		}
	}
}

func getResponseLastModified(resp *http.Response) time.Time {
	h := resp.Header.Get("Last-Modified")
	if h != "" {
		parsed, err := http.ParseTime(h)
		if err == nil {
			return parsed
		}
	}
	return time.Time{}
}
