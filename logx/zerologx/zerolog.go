package zerologx

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/zeromicro/go-zero/core/logx"
)

type (
	ZeroLogWriter struct {
		logger zerolog.Logger
	}
)

func NewZeroLogWriter(logger zerolog.Logger) *ZeroLogWriter {
	return &ZeroLogWriter{logger: logger}
}

func (w *ZeroLogWriter) Alert(v interface{}) {
	w.logger.Error().Msg(fmt.Sprint(v))
}

func (w *ZeroLogWriter) Close() error {
	w.logger.Fatal().Msg("")
	return nil
}

func (w *ZeroLogWriter) Error(v interface{}, fields ...logx.LogField) {
	toZeroLogInterface(w.logger.Error(), fields...).Msgf(fmt.Sprint(v))
}

func (w *ZeroLogWriter) Info(v interface{}, fields ...logx.LogField) {
	toZeroLogInterface(w.logger.Info(), fields...).Msgf(fmt.Sprint(v))
}

func (w *ZeroLogWriter) Severe(v interface{}) {
	w.logger.Fatal().Msg(fmt.Sprint(v))
}

func (w *ZeroLogWriter) Slow(v interface{}, fields ...logx.LogField) {
	toZeroLogInterface(w.logger.Warn(), fields...).Msgf(fmt.Sprint(v))
}

func (w *ZeroLogWriter) Stack(v interface{}) {
	w.logger.Error().Msgf(fmt.Sprint(v))
}

func (w *ZeroLogWriter) Stat(v interface{}, fields ...logx.LogField) {
	toZeroLogInterface(w.logger.Info(), fields...).Msgf(fmt.Sprint(v))
}

func toZeroLogInterface(event *zerolog.Event, fields ...logx.LogField) *zerolog.Event {
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	return event
}
