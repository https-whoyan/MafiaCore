package log

import (
	"context"
	"io"
	"log"
)

const (
	Ldate         = log.Ldate
	Ltime         = log.Ltime
	Lmicroseconds = log.Lmicroseconds
	Llongfile     = log.Llongfile
	Lshortfile    = log.Lshortfile
	LUTC          = log.LUTC
	Lmsgprefix    = log.Lmsgprefix
	LstdFlags     = log.LstdFlags
)

func New(
	w io.Writer,
	prefix string,
	flags int,
) Logger {
	return log.New(
		w,
		prefix,
		flags,
	)
}

type MockCtxLogger struct {
	logger Logger
}

func NewMockCtxLogger(logger Logger) CtxLogger {
	return &MockCtxLogger{
		logger: logger,
	}
}

func (m MockCtxLogger) Info(_ context.Context, v ...any) {
	m.logger.Print(v)
}

func (m MockCtxLogger) Infof(_ context.Context, format string, v ...any) {
	m.logger.Printf(format, v)
}

func (m MockCtxLogger) Infoln(_ context.Context, v ...any) {
	m.logger.Println(v)
}
