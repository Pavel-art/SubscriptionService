package subscriptions

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
	router     *gin.Engine
}

func NewServer(logger *zap.Logger) *Server {
	router := gin.New()

	server := &Server{
		logger: logger,
		router: router,
		httpServer: &http.Server{
			Addr:         ":8080", // Будет переопределено в Start()
			Handler:      router,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	// Устанавливаем middleware при создании
	server.setupMiddleware()
	return server
}

func (s *Server) Start(addr string) error {
	s.httpServer.Addr = addr
	s.logger.Info("Starting HTTP server", zap.String("address", addr))
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server")
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

func (s *Server) setupMiddleware() {
	s.router.Use(
		gin.Recovery(),
		s.loggingMiddleware(), // Теперь метод доступен
	)
}

func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		s.logger.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(start)),
		)
	}
}
