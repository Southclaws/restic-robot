package main

import (
	"os"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	var (
		debug  = os.Getenv("DEBUG")
		config zap.Config
		err    error
	)

	config = zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if runtime.GOOS == "windows" {
		// modify line endings to the Windows-specific version
		config.EncoderConfig.LineEnding = "\r\n"
	}

	if debug != "0" && debug != "" {
		config.Level = zap.NewAtomicLevel()
		config.Level.SetLevel(zap.DebugLevel)
	}

	logger, err = config.Build()
	if err != nil {
		panic(err)
	}
}
