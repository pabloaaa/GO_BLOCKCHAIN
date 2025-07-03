package tests

import (
	"testing"

	"github.com/pabloaaa/GO_BLOCKCHAIN/src"
)

func TestSetLogLevel(t *testing.T) {
	src.SetLogLevel("debug")
	src.SetLogLevel("info")
	src.SetLogLevel("error")
}

func TestInfo(t *testing.T) {
	originalLevel := getLogLevel()

	defer func() {
		src.SetLogLevel(originalLevel)
	}()

	src.SetLogLevel("info")
	src.Info("test info message")
	
	src.SetLogLevel("debug")
	src.Info("test info message debug")
}

func TestDebug(t *testing.T) {
	originalLevel := getLogLevel()

	defer func() {
		src.SetLogLevel(originalLevel)
	}()

	src.SetLogLevel("debug")
	src.Debug("test debug message")
	
	src.SetLogLevel("info")
	src.Debug("test debug message info")
}

func TestError(t *testing.T) {
	src.Error("test error message")
}

func getLogLevel() string {
	return "info"
}