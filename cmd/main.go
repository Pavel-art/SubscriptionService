package main

import (
	"SubscriptionService/internal/subscriptions"
	"SubscriptionService/pkg/db"
	"context"
	"errors"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 1. Инициализация логгера
	logger, err := zap.NewProduction()
	if err != nil {
		panic("не удалось инициализировать логгер: " + err.Error())
	}
	defer logger.Sync()
	logger.Info("Запуск приложения")

	// 2. Загрузка конфигурации БД
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		logger.Fatal("Не задана DB_URL в переменных окружения")
	}

	// 3. Подключение к БД
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbPool, err := db.NewPGXPool(ctx, dbURL, logger)
	if err != nil {
		logger.Fatal("Ошибка подключения к БД", zap.Error(err))
	}
	defer dbPool.Close()
	logger.Info("Успешное подключение к PostgreSQL")

	// 4. Инициализация слоев приложения
	subRepo := subscriptions.NewSubscriptionRepository(dbPool)

	// Инициализация сервера с репозиторием
	apiServer := subscriptions.NewServer(logger)
	apiHandler := subscriptions.NewSubscriptionHandler(logger, subRepo)
	apiHandler.RegisterRoutes(apiServer.GetRouter())

	// 5. Graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("Запуск HTTP сервера", zap.String("port", "8080"))
		if err := apiServer.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Ошибка сервера", zap.Error(err))
			shutdown <- syscall.SIGTERM
		}
	}()

	// Ожидание сигнала завершения
	sig := <-shutdown
	logger.Info("Получен сигнал завершения", zap.String("signal", sig.String()))

	// Завершение работы
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Ошибка при завершении работы сервера", zap.Error(err))
	}

	logger.Info("Приложение корректно завершило работу")
}
