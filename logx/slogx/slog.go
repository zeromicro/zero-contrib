package slogx

import (
	"fmt"
	"log/slog"

	"github.com/zeromicro/go-zero/core/logx"
)

type SlogWriter struct {
	logger *slog.Logger
}

func NewSlogWriter(handler slog.Handler) logx.Writer {
	logger := slog.New(handler)

	return &SlogWriter{
		logger: logger,
	}
}

func (w *SlogWriter) Alert(v interface{}) {
	w.logger.Error(fmt.Sprintf("%v", v))
}

func (w *SlogWriter) Close() error {
	return nil
}

func (w *SlogWriter) Debug(v interface{}, fields ...logx.LogField) {
	w.logger.Debug(fmt.Sprintf("%v", v), toSlogFields(fields...)...)
}

func (w *SlogWriter) Error(v interface{}, fields ...logx.LogField) {
	w.logger.Error(fmt.Sprintf("%v", v), toSlogFields(fields...)...)
}

func (w *SlogWriter) Info(v interface{}, fields ...logx.LogField) {
	w.logger.Info(fmt.Sprintf("%v", v), toSlogFields(fields...)...)
}

func (w *SlogWriter) Severe(v interface{}) {
	w.logger.Error(fmt.Sprintf("%v", v))
}

func (w *SlogWriter) Slow(v interface{}, fields ...logx.LogField) {
	w.logger.Warn(fmt.Sprintf("%v", v), toSlogFields(fields...)...)
}

func (w *SlogWriter) Stack(v interface{}) {
	w.logger.Error(fmt.Sprintf("%v", v))
}

func (w *SlogWriter) Stat(v interface{}, fields ...logx.LogField) {
	w.logger.Info(fmt.Sprintf("%v", v), toSlogFields(fields...)...)
}

func toSlogFields(fields ...logx.LogField) []interface{} {
	slogFields := make([]interface{}, len(fields)*2)
	for i, field := range fields {
		slogFields[i*2] = field.Key
		slogFields[i*2+1] = field.Value
	}

	return slogFields
}
