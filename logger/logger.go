package logger

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log *logrus.Logger

// Init initializes the global logger.
// serviceName is used for the log file name, e.g., "ddns-server.log".
func Init(isInteractive bool, serviceName string) {
	log = logrus.New()

	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		// Fallback or handle error appropriately if LookPath fails
		// For simplicity, we'll try to use the current directory for the log path
		// but this might not be ideal for a service.
		path, _ := os.Getwd()                  // Get current working directory as a fallback
		file = filepath.Join(path, os.Args[0]) // Construct a plausible path
	}

	path, err := filepath.Abs(file)
	if err != nil {
		// Handle error if Abs fails, perhaps log to stderr and exit or use a default path
		// For now, we'll proceed, but this could lead to logs in unexpected places.
		currentDir, _ := os.Getwd()
		path = filepath.Join(currentDir, os.Args[0])
	}

	logFileName := serviceName + ".log"
	logFilePath := filepath.Join(filepath.Dir(path), logFileName)

	logRotate := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    10, // MB
		MaxBackups: 3,
		MaxAge:     7, // days
		Compress:   false,
	}

	log.SetFormatter(&logrus.TextFormatter{TimestampFormat: "2006-01-02 15:04:05"})
	log.SetReportCaller(true)
	log.SetLevel(logrus.InfoLevel)

	if isInteractive {
		writers := []io.Writer{
			logRotate,
			os.Stdout,
		}
		fileAndStdoutWriter := io.MultiWriter(writers...)
		log.SetOutput(fileAndStdoutWriter)
	} else {
		log.SetOutput(logRotate)
	}
}

// L returns the initialized logger instance.
func L() *logrus.Logger {
	if log == nil {
		// Fallback initialization if Init was not called, though Init should always be called first.
		// This basic fallback will log to Stderr.
		basicLog := logrus.New()
		basicLog.SetFormatter(&logrus.TextFormatter{TimestampFormat: "2006-01-02 15:04:05"})
		basicLog.Warn("Logger accessed before initialization. Using basic stderr logger.")
		return basicLog
	}
	return log
}
