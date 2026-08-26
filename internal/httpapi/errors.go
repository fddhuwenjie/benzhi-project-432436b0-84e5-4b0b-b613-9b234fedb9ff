package httpapi

import (
	"mural-biocare/internal/application"
	"mural-biocare/internal/domain"
	"net/http"
	"strings"
)

func statusFor(err error) int {
	if err == nil {
		return http.StatusOK
	}
	code := http.StatusUnprocessableEntity
	if err == domain.ErrInvalid {
		code = http.StatusBadRequest
	}
	if err == domain.ErrConflict {
		code = http.StatusConflict
	}
	if err == domain.ErrForbidden {
		code = http.StatusConflict
	}
	if err == domain.ErrArchived {
		code = http.StatusConflict
	}
	if err == application.ErrNotFound {
		code = http.StatusNotFound
	}
	if err == application.ErrUnauthorized {
		code = http.StatusForbidden
	}
	if strings.Contains(err.Error(), "request_id conflict") {
		code = http.StatusConflict
	}
	return code
}
