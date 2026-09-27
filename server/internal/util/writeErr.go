package util

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type apiError struct {
	status  int
	message any
	err     error
}

func (e *apiError) Error() string {
	return fmt.Sprintf("%s", e.message)
}

type envelope map[string]any

func WriteError(w http.ResponseWriter, r *http.Request, err error, logger *slog.Logger) {
	var apiErr *apiError
	if errors.As(err, &apiErr) {
		logger.Error("api", "error", apiErr.err, "method", r.Method, "uri", r.RequestURI)
		env := envelope{"error": apiErr.message}
		if err := writeJSON(w, env, apiErr.status, nil); err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
func BadRequestErr(message string, err error) *apiError {
	return &apiError{status: http.StatusBadRequest, message: message, err: err}
}
func TimeoutErr(message string, err error) *apiError {
	return &apiError{status: http.StatusGatewayTimeout, message: message, err: err}

}
func UnProcessEntityErr(message string, err error) *apiError {
	return &apiError{status: http.StatusUnprocessableEntity, message: message, err: err}
}
func InternalServerErr(message string, err error) *apiError {
	return &apiError{status: http.StatusInternalServerError, message: message, err: err}
}
func NotFounErr(message string, err error) *apiError {
	return &apiError{status: http.StatusNotFound, message: message, err: err}
}
func ValidationErr(message map[string]string) *apiError {
	return &apiError{status: http.StatusUnprocessableEntity, message: message, err: nil}
}
func UnsupportedMediaTypeErr(message string, err error) *apiError {

	return &apiError{status: http.StatusUnsupportedMediaType, message: message, err: err}
}
func UnauthorizedErr(message string, err error) *apiError {
	return &apiError{status: http.StatusUnauthorized, message: message, err: err}
}
func ConflictErr(message string, err error) *apiError {
	return &apiError{status: http.StatusConflict, message: message, err: err}
}
func ForbiddenErr(message string, err error) *apiError {
	return &apiError{status: http.StatusForbidden, message: message, err: err}
}
func BadGateWayErr(message string, err error) *apiError {
	return &apiError{status: http.StatusBadGateway, message: message, err: err}
}
