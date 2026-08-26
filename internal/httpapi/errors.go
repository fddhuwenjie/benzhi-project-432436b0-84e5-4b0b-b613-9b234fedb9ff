package httpapi

import "net/http"

func statusFor(err error) int {
	if err == nil {
		return http.StatusOK
	}
	return http.StatusUnprocessableEntity
}
