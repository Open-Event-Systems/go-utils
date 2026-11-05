package httputils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
)

type JSONErrorObject struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Errors  []any  `json:"errors,omitempty"`
}

type JSONErrorResponseBody struct {
	Error JSONErrorObject `json:"error"`
}

var JSONErrorResponseWriter = ErrorResponseWriterFunc(func(w http.ResponseWriter, status int, message string) {
	body := JSONErrorResponseBody{
		Error: JSONErrorObject{
			Code:    status,
			Message: message,
		},
	}

	buf := bytes.Buffer{}

	enc := json.NewEncoder(&buf)
	err := enc.Encode(body)

	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.WriteHeader(status)
	buf.WriteTo(w)
})

var JSONErrorHandler = ErrorHandler{
	ErrorResponseWriter: JSONErrorResponseWriter,
}
