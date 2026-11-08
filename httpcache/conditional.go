package httpcache

import (
	"net/http"
)

// Add conditional request headers to a [http.Header]
func (ri ResponseInfo) AddConditionalRequestHeaders(h http.Header) {
	if ri.ETag != "" {
		h.Set("If-None-Match", ri.ETag)
	}

	if !ri.LastModified.IsZero() {
		h.Set("If-Modified-Since", ri.LastModified.UTC().Format(http.TimeFormat))
	}
}
