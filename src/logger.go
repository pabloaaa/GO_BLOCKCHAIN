package src

import (
	"log"
	"os"
	"time"
)

var (
	infoLogger  *log.Logger
	debugLogger *log.Logger
	errorLogger *log.Logger
	logLevel    string
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

func init() {
	log.SetFlags(0) // Disable default timestamp
	infoLogger = log.New(os.Stdout, "", log.Lmsgprefix)
	debugLogger = log.New(os.Stdout, "", log.Lmsgprefix)
	errorLogger = log.New(os.Stderr, "", log.Lmsgprefix)
	logLevel = "info" // Default log level
}

func SetLogLevel(level string) {
	logLevel = level
}

func logWithTimestamp(logger *log.Logger, levelColor string, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	coloredTimestamp := colorGreen + timestamp + colorReset
	logger.Printf("%s %s%v%s", coloredTimestamp, levelColor, message, colorReset)
}

func Info(message string) {
	if logLevel == "info" || logLevel == "debug" {
		logWithTimestamp(infoLogger, colorGreen+"INFO: "+colorReset, message)
	}
}

func Debug(message string) {
	if logLevel == "debug" {
		logWithTimestamp(debugLogger, colorYellow+"DEBUG: "+colorReset, message)
	}
}

func Error(message string) {
	logWithTimestamp(errorLogger, colorRed+"ERROR: "+colorReset, message)
}
