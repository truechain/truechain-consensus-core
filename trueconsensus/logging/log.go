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
	"github.com/sirupsen/logrus"
)

const (
	// LogInfo checks if the log level is INFO
	LogInfo string = "INFO"
	// LogError checks if the log level is ERROR
	LogError string = "ERROR"
	// LogWarning checks if the log level is WARNING
	LogWarning string = "WARNING"
)

// Logger defines the interface for logging used by gophoton-core internally
type Logger interface {
	logrus.FieldLogger
	LoggerWith(logrus.Fields) Logger
}

var rootLogger Logger

// InitRootLoger initialize root logger
func InitRootLoger() {
	rootLogger = GetLogger()
}

// pbftLog prepends the passed module to the log message.
func pbftLog(logLevel string, module string, format string, v ...interface{}) {
	if logLevel == LogInfo {
		rootLogger.WithField("Module", module).Infof(format, v...)
	}
	if logLevel == LogError {
		rootLogger.WithField("Module", module).Errorf(format, v...)
	}
	if logLevel == LogWarning {
		rootLogger.WithField("Module", module).Warningf(format, v...)
	}
}
