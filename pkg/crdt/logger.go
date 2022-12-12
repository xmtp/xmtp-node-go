package crdt

import (
	"fmt"

	"go.uber.org/zap"
)

type logger struct {
	zap *zap.Logger
}

func (l *logger) Debug(args ...interface{}) {
	l.zap.Debug(fmt.Sprint(args...))
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.zap.Debug(fmt.Sprintf(format, args...))
}

func (l *logger) Error(args ...interface{}) {
	l.zap.Error(fmt.Sprint(args...))
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.zap.Error(fmt.Sprintf(format, args...))
}

func (l *logger) Fatal(args ...interface{}) {
	l.zap.Fatal(fmt.Sprint(args...))
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.zap.Fatal(fmt.Sprintf(format, args...))
}

func (l *logger) Info(args ...interface{}) {
	l.zap.Info(fmt.Sprint(args...))
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.zap.Info(fmt.Sprintf(format, args...))
}

func (l *logger) Panic(args ...interface{}) {
	l.zap.Panic(fmt.Sprint(args...))
}

func (l *logger) Panicf(format string, args ...interface{}) {
	l.zap.Panic(fmt.Sprintf(format, args...))
}

func (l *logger) Warn(args ...interface{}) {
	l.zap.Warn(fmt.Sprint(args...))
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.zap.Warn(fmt.Sprintf(format, args...))
}
