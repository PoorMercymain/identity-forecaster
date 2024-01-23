package logger

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type logger struct {
	file   string
	once   *sync.Once
	logger *zap.SugaredLogger
}

const (
	defaultLogfilePath = "logfile.log"
	stdoutStr          = "stdout"
)

var log = &logger{file: defaultLogfilePath, once: &sync.Once{}}

func SetLogfilePath(path string) {
	log.file = path
}

func Logger() *zap.SugaredLogger {
	var err error
	log.once.Do(func() {
		cfg := zap.NewProductionConfig()
		cfg.OutputPaths = []string{log.file, stdoutStr}

		var logger *zap.Logger
		logger, err = cfg.Build()

		log.logger = logger.Sugar()
	})

	if err != nil {
		panicStr := fmt.Sprintf("cannot proceed: couldn't initialize logger because of an error - %v", err)
		panic(panicStr)
	}

	return log.logger
}
