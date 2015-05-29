package camillo

import (
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"
)

// Logger is a middleware handler that logs the request as it goes in and the response as it goes out.
type Logger struct {
	// Logger inherits from log.Logger used to log messages with the Logger middleware
	*log.Logger
}

// NewLogger returns a new Logger instance
func NewLogger() *Logger {
	return &Logger{log.New(os.Stdout, "[camillo] ", 0)}
}

func (l *Logger) ServeHTTP(ctx context.Context, rw http.ResponseWriter, r *http.Request, next NextFunc) {
	start := time.Now()
	l.Printf("Started %s %s", r.Method, r.URL.Path)

	next(ctx, rw, r)

	res := rw.(ResponseWriter)
	l.Printf("Completed %v %s in %v", res.Status(), http.StatusText(res.Status()), time.Since(start))
}
