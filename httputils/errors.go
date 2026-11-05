package httputils

import (
	"net/http"

	"github.com/go-chi/httplog/v3"
)

type ErrorResponseWriter interface {
	// Write an error response with the given status code and message.
	WriteResponseMessage(w http.ResponseWriter, status int, message string)
}

type ErrorResponseWriterFunc func(w http.ResponseWriter, status int, message string)

type ErrorHandler struct {
	ErrorResponseWriter ErrorResponseWriter
}

func (f ErrorResponseWriterFunc) WriteResponseMessage(w http.ResponseWriter, status int, message string) {
	f(w, status, message)
}

var DefaultErrorResponseWriter = ErrorResponseWriterFunc(func(w http.ResponseWriter, status int, message string) {
	http.Error(w, message, status)
})

// Return a new [ErrorHandler].
func NewErrorHandler(rw ErrorResponseWriter) ErrorHandler {
	return ErrorHandler{
		ErrorResponseWriter: rw,
	}
}

// Write an error response with the given status code and message.
func (h ErrorHandler) WriteResponseMessage(w http.ResponseWriter, status int, message string) {
	h.ErrorResponseWriter.WriteResponseMessage(w, status, message)
}

// Write an error response with the given status code.
func (h ErrorHandler) WriteResponse(w http.ResponseWriter, status int) {
	h.ErrorResponseWriter.WriteResponseMessage(w, status, http.StatusText(status))
}

// Write a not found error response.
func (h ErrorHandler) NotFound(w http.ResponseWriter) {
	h.WriteResponse(w, http.StatusNotFound)
}

// If err is not nil, write an error response to w and return true.
func (h ErrorHandler) CheckError(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}

	ctx := r.Context()
	httplog.SetError(ctx, err)

	h.WriteResponse(w, http.StatusInternalServerError)

	return true
}

var DefaultErrorHandler = ErrorHandler{
	ErrorResponseWriter: DefaultErrorResponseWriter,
}
