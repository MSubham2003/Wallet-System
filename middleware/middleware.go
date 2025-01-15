package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs HTTP requests with status codes and execution time
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Use a ResponseWriter wrapper to capture the status code
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(rec, r)

		// Log the request details
		log.Printf("%s %s %s %d %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			rec.statusCode,
			time.Since(start),
		)
	})
}

// responseRecorder is a wrapper to capture the status code
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *responseRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}
