package httputils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
)

type JSONDataResponseBody struct {
	Data any `json:"data"`
}

type JSONDataRequestBody[T any] struct {
	Data T `json:"data"`
}

type JSONRequestReader interface {
	// Decode a JSON request body into val. If an error occurs, responds to the request with an error and returns true.
	ReadBody(w http.ResponseWriter, r *http.Request, val any) bool
}

type JSONRequestReaderFunc func(w http.ResponseWriter, r *http.Request, val any) bool

func (f JSONRequestReaderFunc) ReadBody(w http.ResponseWriter, r *http.Request, val any) bool {
	return f(w, r, val)
}

func NewJSONRequestReader(handler ErrorHandler) JSONRequestReader {
	return JSONRequestReaderFunc(func(w http.ResponseWriter, r *http.Request, val any) bool {
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(val)
		if err != nil {
			handler.WriteResponse(w, http.StatusBadRequest)
			return true
		} else {
			return false
		}
	})
}

var DefaultJSONRequestReader = NewJSONRequestReader(JSONErrorHandler)

// Send a JSON response body with the given status. Set content type and length headers.
func WriteJSONResponseStatus(w http.ResponseWriter, status int, val any) error {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	err := enc.Encode(val)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.WriteHeader(status)
	_, err = buf.WriteTo(w)
	return err
}

// Send a JSON response body. Set content type and length headers.
func WriteJSONResponse(w http.ResponseWriter, val any) error {
	return WriteJSONResponseStatus(w, http.StatusOK, val)
}
