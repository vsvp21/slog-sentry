# slogsentry

Package to integrate slog and sentry

To initialize with io.MultiWriter:
```go
package main

import (
	"io"
	"log/slog"
	"os"

	"github.com/getsentry/sentry-go"
	slogsentry "github.com/vsvp21/slog-sentry"
)

func main() {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:           "dsn",
		EnableTracing: true,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,
	}); err != nil {
		return err
	}

	w := slogsentry.NewWriter(sentry.CurrentHub().Client(), time.Millisecond*500)

	return slog.New(slog.NewJSONHandler(io.MultiWriter(os.Stdout, w), opts)).With(slog.String("service", "servicename"))
}
```
