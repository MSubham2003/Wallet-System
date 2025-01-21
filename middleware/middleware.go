package middleware

import (
	"fmt"
	"net/http"
)

// LoggingMiddleware logs HTTP requests with status codes, execution time, and color-coded status
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// start := time.Now()

		// Use a ResponseWriter wrapper to capture the status code
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(rec, r)

		// Format log entry
		// logEntry := fmt.Sprintf(
		// 	"\t\tStatus: %s %-20s URI:-%-15s\t %-10s\t %s",
		// 	colorizeStatusCode(rec.statusCode), // Color-coded status
		// 	r.Method,
		// 	r.RequestURI,
		// 	// r.RemoteAddr,
		// 	time.Since(start),
		// 	resetColor(),
		// )

		// // Log the request details
		// log.Println(logEntry)
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

// colorizeStatusCode adds color based on status code
func colorizeStatusCode(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return fmt.Sprintf("\033[32m%d\033[0m", statusCode) // Green for 2xx
	case statusCode >= 400 && statusCode < 500:
		return fmt.Sprintf("\033[33m%d\033[0m", statusCode) // Yellow for 4xx
	case statusCode >= 500:
		return fmt.Sprintf("\033[31m%d\033[0m", statusCode) // Red for 5xx
	default:
		return fmt.Sprintf("%d", statusCode) // Default color
	}
}

// resetColor resets ANSI color
func resetColor() string {
	return "\033[0m"
}
