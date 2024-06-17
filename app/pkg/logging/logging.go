package logging

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"strings"
)

var e *logrus.Entry

type Logger struct {
	*logrus.Entry
}

func GetLogger() *Logger {
	return &Logger{e}
}

func (l *Logger) GetLoggerWithField(key string, value interface{}) *Logger {
	return &Logger{l.Entry.WithField(key, value)}
}

type writerHook struct {
	Writers   []io.Writer
	LogLevels []logrus.Level
}

func (hook *writerHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	for _, w := range hook.Writers {
		_, err = w.Write([]byte(line))
	}
	return err
}

func (hook *writerHook) Levels() []logrus.Level {
	return hook.LogLevels
}

func InitLogger() {
	l := logrus.New()
	l.SetReportCaller(true)

	customFormatter := &CustomFormatter{}
	l.SetFormatter(customFormatter)

	err := os.MkdirAll("logs", 0755)

	if err != nil || os.IsExist(err) {
		panic("can't create log dir. logging to file is not configured")
	} else {
		logFile, err := os.OpenFile("logs/all.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
		if err != nil {
			panic(fmt.Sprintf("[Error]: %s", err))
		}

		l.SetOutput(io.Discard)
		l.AddHook(&writerHook{
			Writers:   []io.Writer{logFile, os.Stdout},
			LogLevels: logrus.AllLevels,
		})
	}

	l.SetLevel(logrus.TraceLevel)

	e = logrus.NewEntry(l)
}

type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	level := strings.ToUpper(entry.Level.String())

	funcName := ""
	file := ""
	if entry.Caller != nil {
		funcName = fmt.Sprintf(" [%s()]", entry.Caller.Function)
		file = fmt.Sprintf("- %s:%d", path.Base(entry.Caller.File), entry.Caller.Line)
	}

	//formatted := fmt.Sprintf("%s \u001B[%dm%s\u001B[0m %s %s %s \n",
	//	timestamp, getColorByLevel(entry.Level), level, entry.Message, funcName, file)
	formatted := fmt.Sprintf("%s %s %s %s %s\n",
		timestamp, level, entry.Message, funcName, file)
	return []byte(formatted), nil
}

func getColorByLevel(level logrus.Level) int {
	switch level {
	case logrus.TraceLevel:
		return 36
	case logrus.DebugLevel:
		return 32
	case logrus.InfoLevel:
		return 34
	case logrus.WarnLevel:
		return 33
	case logrus.ErrorLevel:
		return 31
	case logrus.FatalLevel, logrus.PanicLevel:
		return 35
	default:
		return 0
	}
}
