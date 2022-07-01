package zerologx

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/zeromicro/go-zero/core/logx"
)

type (
	ZeroLogWrite struct {
		logger zerolog.Logger
	}
)

func NewZeroLogWrite(logger zerolog.Logger) *ZeroLogWrite {
	return &ZeroLogWrite{logger: logger}
}

func (w *ZeroLogWrite) Alert(v interface{}) {
	w.logger.Error().Msg(fmt.Sprint(v))
}

func (w *ZeroLogWrite) Close() error {
	w.logger.Fatal().Msg("")
	return nil
}

func (w *ZeroLogWrite) Error(v interface{}, fields ...logx.LogField) {
	toZeroLogInterface(w.logger.Error(), fields...).Msgf(fmt.Sprint(v))
}

func (w *ZeroLogWrite) Info(v interface{}, fields ...logx.LogField) {
	toZeroLogInterface(w.logger.Info(), fields...).Msgf(fmt.Sprint(v))
}

func (w *ZeroLogWrite) Severe(v interface{}) {
	w.logger.Fatal().Msg(fmt.Sprint(v))
}

func (w *ZeroLogWrite) Slow(v interface{}, fields ...logx.LogField) {
	toZeroLogInterface(w.logger.Warn(), fields...).Msgf(fmt.Sprint(v))
}

func (w *ZeroLogWrite) Stack(v interface{}) {
	w.logger.Error().Msgf(fmt.Sprint(v))
}

func (w *ZeroLogWrite) Stat(v interface{}, fields ...logx.LogField) {
	toZeroLogInterface(w.logger.Info(), fields...).Msgf(fmt.Sprint(v))
}

func toZeroLogInterface(event *zerolog.Event, fields ...logx.LogField) *zerolog.Event {
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	return event
}
