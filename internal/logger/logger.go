package logger

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"log/slog"

	"github.com/fatih/color"
)

type CustomHandlerOpts struct {
	SlogOpts slog.HandlerOptions
}

type CustomHandler struct {
	slog.Handler
	l *log.Logger
}

func (ch *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String() + ":"
	switch r.Level {
	case slog.LevelDebug:
		level = color.WhiteString(level)
	case slog.LevelInfo:
		level = color.GreenString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	default:
		level = color.HiWhiteString(level)
	}

	timeStr := r.Time.Format("[15:04:05]")
	message := color.HiWhiteString(r.Message)

	if r.NumAttrs() == 0 {
		ch.l.Println(timeStr, level, message)
		return nil
	}

	fields := make(map[string]interface{}, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})

	j, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return err
	}

	ch.l.Println(timeStr, level, message, color.WhiteString(string(j)))
	return nil
}

func NewCustomHandler(out io.Writer, opts CustomHandlerOpts) *CustomHandler {
	return &CustomHandler{
		Handler: slog.NewJSONHandler(out, &opts.SlogOpts),
		l:       log.New(out, "", 0),
	}
}
