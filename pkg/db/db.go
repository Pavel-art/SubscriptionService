package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"time"
)

func NewPGXPool(ctx context.Context, connString string, logger *zap.Logger) (*pgxpool.Pool, error) {

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		logger.Error("Ошибка парсинга строки подключения", zap.Error(err))
		return nil, err
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	logger.Debug("Настройки пула соединений",
		zap.Int("max_conns", int(config.MaxConns)),
		zap.Int("min_conns", int(config.MinConns)),
	)

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		logger.Error("Ошибка создания пула соединений", zap.Error(err))
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		logger.Error("Проверка подключения не пройдена", zap.Error(err))
		return nil, err
	}

	logger.Info("Пул соединений к PostgreSQL успешно создан",
		zap.String("db_host", config.ConnConfig.Host),
		zap.String("db_name", config.ConnConfig.Database),
	)

	return pool, nil
}
