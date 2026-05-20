package middleware

import (
	"net/http"
	"time"
)

func RequestTimeout(next http.Handler) http.Handler {
	return http.TimeoutHandler(next, 60*time.Second, `{"error":"request timeout"}`)
}
