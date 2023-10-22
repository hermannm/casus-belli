package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"

	"hermannm.dev/devlog"
	"hermannm.dev/wrap"
)

func Initialize() {
	logger := slog.New(devlog.NewHandler(os.Stdout, &devlog.Options{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
}

func Info(msg string, metadata ...slog.Attr) {
	log(slog.LevelInfo, msg, metadata...)
}

func Infof(format string, args ...any) {
	log(slog.LevelInfo, fmt.Sprintf(format, args...))
}

func Warn(msg string, metadata ...slog.Attr) {
	log(slog.LevelWarn, msg, metadata...)
}

func Warnf(format string, args ...any) {
	log(slog.LevelWarn, fmt.Sprintf(format, args...))
}

func Error(err error, msg string, metadata ...slog.Attr) {
	if err == nil {
		log(slog.LevelError, msg, metadata...)
	} else {
		if msg != "" {
			err = wrap.Error(err, msg)
		}

		log(slog.LevelError, err.Error(), metadata...)
	}
}

func Errorf(err error, format string, args ...any) {
	if err == nil {
		log(slog.LevelError, fmt.Sprintf(format, args...))
	} else {
		log(slog.LevelError, wrap.Errorf(err, format, args...).Error())
	}
}

func Debug(msg string, metadata ...slog.Attr) {
	log(slog.LevelDebug, msg, metadata...)
}

func Debugf(format string, args ...any) {
	log(slog.LevelDebug, fmt.Sprintf(format, args...))
}

func log(level slog.Level, msg string, metadata ...slog.Attr) {
	logger := slog.Default()
	if !logger.Enabled(context.Background(), level) {
		return
	}

	// Follows the example from the slog package of how to properly wrap its functions:
	// https://pkg.go.dev/golang.org/x/exp/slog#hdr-Wrapping_output_methods
	var callers [1]uintptr
	// Skips 3, because we want to skip:
	// - the call to Callers
	// - the call to log (this function)
	// - the call to the public log function that uses this function
	runtime.Callers(3, callers[:])

	record := slog.NewRecord(time.Now(), level, msg, callers[0])
	record.AddAttrs(metadata...)

	_ = logger.Handler().Handle(context.Background(), record)
}
