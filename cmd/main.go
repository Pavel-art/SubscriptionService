package main

import (
	_ "SubscriptionService/docs"
	"SubscriptionService/internal/subscriptions"
	"SubscriptionService/pkg/db"

	"context"
	"errors"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title Subscription API
// @version 1.0
// @description API для управления онлайн-подписками
// @host localhost:8080
// @BasePath /api/v1
// @schemes http
// @contact.name API Support
// @contact.email support@subscription.com
func main() {
	//  Инициализация логгера
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("не удалось инициализировать логгер: " + err.Error())
	}
	defer logger.Sync()
	logger.Info("Запуск приложения")

	// Загрузка конфигурации из .env файла
	if err := godotenv.Load(); err != nil {
		logger.Fatal("Ошибка загрузки .env файла", zap.Error(err))
	}
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		logger.Fatal("Не задана DB_URL в переменных окружения")
	}

	// Подключение к базе данных и создание контекста
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Освобождаем ресурсы

	dbPool, err := db.NewPGXPool(ctx, dbURL, logger)
	if err != nil {
		logger.Fatal("Ошибка подключения к БД", zap.Error(err))
	}
	defer dbPool.Close() // Закрываем соединение с БД при завершении
	logger.Info("Успешное подключение к PostgreSQL")

	// Инициализация репозитория
	subRepo := subscriptions.NewSubscriptionRepository(dbPool, logger)

	//Создание сервера и обработчиков, Регистрация маршрутов API
	apiServer := subscriptions.NewServer(logger)
	apiHandler := subscriptions.NewSubscriptionHandler(logger, subRepo)
	apiHandler.RegisterRoutes(apiServer.GetRouter())

	//Настройка graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("Запуск HTTP сервера", zap.String("port", "8080"))
		if err := apiServer.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Ошибка сервера", zap.Error(err))
			shutdown <- syscall.SIGTERM
		}
	}()

	sig := <-shutdown
	logger.Info("Получен сигнал завершения", zap.String("signal", sig.String()))

	//Graceful shutdown сервера
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Ошибка при завершении работы сервера", zap.Error(err))
	}

	logger.Info("Приложение корректно завершило работу")
}
