/*
Copyright (c) 2018 TrueChain Foundation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

// LogFor retrieves a logger for the context
func LogFor() Logger {
	return &logger{entry: logrus.NewEntry(setContext())}
}

type logger struct {
	entry *logrus.Entry
}

// LoggerIdentity stores context for logger
type LoggerIdentity struct {
	RequestID string
	UserID    string
	SessionID string
}

func (l *logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.entry.WithField(key, value)
}
func (l *logger) WithFields(fields logrus.Fields) *logrus.Entry { return l.entry.WithFields(fields) }
func (l *logger) WithError(err error) *logrus.Entry             { return l.entry.WithError(err) }

func (l *logger) Debugf(format string, args ...interface{})   { l.entry.Debugf(format, args...) }
func (l *logger) Infof(format string, args ...interface{})    { l.entry.Infof(format, args...) }
func (l *logger) Printf(format string, args ...interface{})   { l.entry.Printf(format, args...) }
func (l *logger) Warnf(format string, args ...interface{})    { l.entry.Warnf(format, args...) }
func (l *logger) Warningf(format string, args ...interface{}) { l.entry.Warningf(format, args...) }
func (l *logger) Errorf(format string, args ...interface{})   { l.entry.Errorf(format, args...) }
func (l *logger) Fatalf(format string, args ...interface{})   { l.entry.Fatalf(format, args...) }
func (l *logger) Panicf(format string, args ...interface{})   { l.entry.Panicf(format, args...) }

func (l *logger) Debug(args ...interface{})   { l.entry.Debug(args...) }
func (l *logger) Info(args ...interface{})    { l.entry.Info(args...) }
func (l *logger) Print(args ...interface{})   { l.entry.Print(args...) }
func (l *logger) Warn(args ...interface{})    { l.entry.Warn(args...) }
func (l *logger) Warning(args ...interface{}) { l.entry.Warning(args...) }
func (l *logger) Error(args ...interface{})   { l.entry.Error(args...) }
func (l *logger) Fatal(args ...interface{})   { l.entry.Fatal(args...) }
func (l *logger) Panic(args ...interface{})   { l.entry.Panic(args...) }

func (l *logger) Debugln(args ...interface{})   { l.entry.Debugln(args...) }
func (l *logger) Infoln(args ...interface{})    { l.entry.Infoln(args...) }
func (l *logger) Println(args ...interface{})   { l.entry.Println(args...) }
func (l *logger) Warnln(args ...interface{})    { l.entry.Warnln(args...) }
func (l *logger) Warningln(args ...interface{}) { l.entry.Warningln(args...) }
func (l *logger) Errorln(args ...interface{})   { l.entry.Errorln(args...) }
func (l *logger) Fatalln(args ...interface{})   { l.entry.Fatalln(args...) }
func (l *logger) Panicln(args ...interface{})   { l.entry.Panicln(args...) }

func setContext() *logrus.Logger {

	logContext := logrus.Logger{
		Out:       os.Stdout,
		Formatter: new(logrus.TextFormatter),
		Level:     logrus.InfoLevel}
	return &logContext
}

func (l *logger) LoggerWith(fields logrus.Fields) Logger {
	return &logger{entry: l.entry.WithFields(fields)}
}

// SetLogProps creates a map of the log properties
func SetLogProps() map[string]string {
	logProps := make(map[string]string)
	logProps["truechain"] = "pbft"
	return logProps
}

// GetLogger returns the logger
func GetLogger() Logger {
	fields := make(logrus.Fields)
	logProperties := SetLogProps()
	for k, v := range logProperties {
		fields[k] = v
	}
	l := LogFor()
	if len(fields) > 0 {
		l = l.LoggerWith(fields)
	}
	return l
}
