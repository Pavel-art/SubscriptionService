package logging

import "go.uber.org/zap"

func NewLogger(level string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()

	logLevel := zap.InfoLevel
	if err := logLevel.UnmarshalText([]byte(level)); err != nil {
		return nil, err
	}
	cfg.Level = zap.NewAtomicLevelAt(logLevel)

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
