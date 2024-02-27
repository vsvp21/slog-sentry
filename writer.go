package slogsentry

import (
	"github.com/buger/jsonparser"
	"github.com/getsentry/sentry-go"
	"time"
	"unsafe"
)

const (
	MessageKey = "message"
	ErrorKey   = "error"
)

var levelsMapping = map[string]sentry.Level{
	"DEBUG": sentry.LevelDebug,
	"INFO":  sentry.LevelInfo,
	"WARN":  sentry.LevelWarning,
	"ERROR": sentry.LevelError,
}

var enabledLevels = map[sentry.Level]struct{}{
	sentry.LevelError: {},
}

func NewWriter(client *sentry.Client, flushTimeout time.Duration) *Writer {
	return &Writer{
		client:       client,
		flushTimeout: flushTimeout,
	}
}

type Writer struct {
	client       *sentry.Client
	flushTimeout time.Duration
}

func (w *Writer) Write(data []byte) (int, error) {
	event, err := w.parseEvent(data)
	if err != nil {
		return 0, err
	}

	if event == nil {
		return 0, nil
	}

	w.client.CaptureEvent(event, nil, nil)

	return len(data), nil
}

func (w *Writer) Close() error {
	sentry.Flush(w.flushTimeout)
	return nil
}

func (w *Writer) parseEvent(data []byte) (*sentry.Event, error) {
	lvlStr, err := jsonparser.GetUnsafeString(data, "level")
	if err != nil {
		return nil, err
	}

	sentryLvl := levelsMapping[lvlStr]
	if _, enabled := enabledLevels[sentryLvl]; !enabled {
		return nil, nil
	}

	event := sentry.Event{
		Timestamp: time.Now().UTC(),
		Level:     sentryLvl,
		Logger:    "slog",
	}

	err = jsonparser.ObjectEach(data, func(key, value []byte, vt jsonparser.ValueType, offset int) error {
		switch string(key) {
		case MessageKey:
			event.Message = bytesToStrUnsafe(value)
		case ErrorKey:
			event.Exception = append(event.Exception, sentry.Exception{
				Value:      bytesToStrUnsafe(value),
				Stacktrace: stackTrace(),
			})
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func bytesToStrUnsafe(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// Ignore if your IDE shows an error here; it's a false positive.
	p := unsafe.SliceData(data)
	return unsafe.String(p, len(data))
}

func stackTrace() *sentry.Stacktrace {
	const (
		currentModule = "feature/internal/infrastructure/logger"
	)

	st := sentry.NewStacktrace()

	threshold := len(st.Frames) - 1
	for ; threshold > 0 && st.Frames[threshold].Module == currentModule; threshold-- {
	}

	st.Frames = st.Frames[:threshold+1]

	return st
}
