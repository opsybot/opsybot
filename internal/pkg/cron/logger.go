package cron

import "log/slog"

type schedulerLogger struct{ log *slog.Logger }

func (l schedulerLogger) Debug(msg string, args ...any) { l.log.Debug(msg, args...) }
func (l schedulerLogger) Info(msg string, args ...any)  { l.log.Info(msg, args...) }
func (l schedulerLogger) Warn(msg string, args ...any)  { l.log.Warn(msg, args...) }
func (l schedulerLogger) Error(msg string, args ...any) { l.log.Error(msg, args...) }
