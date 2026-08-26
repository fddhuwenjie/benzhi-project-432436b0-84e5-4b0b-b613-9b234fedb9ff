package httpapi

import "net/http"

func actor(r *http.Request) string { return r.Header.Get("X-Actor") }
