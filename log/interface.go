package log

import (
	"context"
	"io"
)

type Logger interface {
	SetOutput(w io.Writer)
	Output(calldepth int, s string) error
	Print(v ...any)
	Printf(format string, v ...any)
	Println(v ...any)
	Fatal(v ...any)
	Fatalf(format string, v ...any)
	Fatalln(v ...any)
	Panic(v ...any)
	Panicf(format string, v ...any)
	Panicln(v ...any)
	Flags() int
	SetFlags(flag int)
	Prefix() string
	SetPrefix(prefix string)
	Writer() io.Writer
}

type CtxLogger interface {
	Info(ctx context.Context, v ...any)
	Infof(ctx context.Context, format string, v ...any)
	Infoln(ctx context.Context, v ...any)
}

type CommonLogger interface {
	Logger
	CtxLogger
}
