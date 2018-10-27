package main

import (
	"os"

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

	if debug != "0" && debug != "" {
		config.Level = zap.NewAtomicLevel()
		config.Level.SetLevel(zap.DebugLevel)
	}

	logger, err = config.Build()
	if err != nil {
		panic(err)
	}
}
