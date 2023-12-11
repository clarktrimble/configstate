package handmock

import (
	"context"
	"fmt"
)

type Logger struct {
	Logged    []map[string]any
	CtxFields map[string]any
}

func NewLogger() (logger *Logger) {

	logger = &Logger{
		CtxFields: map[string]any{},
		Logged:    []map[string]any{},
	}
	return
}

func (ml *Logger) Info(ctx context.Context, msg string, kv ...any) {
	ml.log(ctx, msg, kv)
}

func (ml *Logger) Error(ctx context.Context, msg string, err error, kv ...any) {
	ml.log(ctx, msg, append(kv, "error", fmt.Sprintf("%+v", err)))
}

func (ml *Logger) WithFields(ctx context.Context, kv ...any) context.Context {

	for key, val := range strKey(kv) {
		ml.CtxFields[key] = val
	}

	return ctx
}

// unexported

func (ml *Logger) log(ctx context.Context, msg string, kv []any) {

	line := strKey(kv)

	for key, val := range ml.CtxFields {
		line[key] = val
	}

	line["msg"] = msg

	ml.Logged = append(ml.Logged, line)
}

func strKey(pairs []any) (keyval map[string]any) {

	keyval = map[string]any{}

	for i := 0; i < len(pairs)-1; i = i + 2 {
		key := pairs[i].(string)
		val := pairs[i+1]

		keyval[key] = val
	}

	return
}
