package error

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewCustomError(code string, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    err.Error(),
		StatusCode: mapCodeToStatus(code),
		Err:        err,
	}
}

func FromError(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    err.Error(),
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

func (e *AppError) WriteJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.StatusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"code":    e.Code,
		"message": e.Message,
	})
}

func mapCodeToStatus(code string) int {
	switch code {
	case "ERR_NOT_FOUND":
		return http.StatusNotFound
	case "ERR_VALIDATION":
		return http.StatusBadRequest
	case "ERR_UNAUTHORIZED":
		return http.StatusUnauthorized
	case "ERR_FORBIDDEN":
		return http.StatusForbidden
	case "ERR_CONFLICT":
		return http.StatusConflict
	case "ERR_UNSUPPORTED":
		return http.StatusNotImplemented
	case "ERR_MISSING_FIELD":
		return http.StatusBadRequest
	case "ERR_INVALID_INPUT":
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func EncodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	code := http.StatusInternalServerError
	message := err.Error()

	if errors.Is(err, context.DeadlineExceeded) {
		code = http.StatusRequestTimeout
		message = "Request timed out"
	} else if message == "id is required" || message == "id and message are required" {
		code = http.StatusBadRequest
	}

	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
