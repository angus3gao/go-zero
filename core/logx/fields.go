package logx

import (
	"context"
	"sync"
	"sync/atomic"
)

var (
	globalFields     atomic.Value
	globalFieldsLock sync.Mutex
)

type fieldsKey struct{}

type LogFields struct {
	LogFields []LogField
}

// AddGlobalFields adds global fields.
func AddGlobalFields(fields ...LogField) {
	globalFieldsLock.Lock()
	defer globalFieldsLock.Unlock()

	old := globalFields.Load()
	if old == nil {
		globalFields.Store(append([]LogField(nil), fields...))
	} else {
		globalFields.Store(append(old.([]LogField), fields...))
	}
}

// ContextWithFields returns a new context with the given fields.
func ContextWithFields(ctx context.Context, fields ...LogField) context.Context {
	if val := ctx.Value(fieldsKey{}); val != nil {
		if arr, ok := val.([]LogField); ok {
			allFields := make([]LogField, 0, len(arr)+len(fields))
			allFields = append(allFields, arr...)
			allFields = append(allFields, fields...)
			return context.WithValue(ctx, fieldsKey{}, allFields)
		}
	}

	return context.WithValue(ctx, fieldsKey{}, fields)
}

// WithFields returns a new logger with the given fields.
// deprecated: use ContextWithFields instead.
func WithFields(ctx context.Context, fields ...LogField) context.Context {
	return ContextWithFields(ctx, fields...)
}
func NewLogFields(logFields ...LogField) *LogFields {
	fields := LogFields{}
	fields.LogFields = make([]LogField, 0, len(logFields))
	fields.LogFields = append(fields.LogFields, logFields...)
	return &fields
}

func (lf *LogFields) AddField(logField LogField) {
	lf.LogFields = append(lf.LogFields, logField)
}

func (lf *LogFields) GetFields() []LogField {
	return lf.LogFields
}

func ParseFields(args ...any) ([]any, *LogFields) {
	if len(args) > 0 {
		tail := len(args) - 1
		if fields, ok := args[tail].(*LogFields); ok {
			return args[:tail], fields
		}
	}

	return args, nil
}
