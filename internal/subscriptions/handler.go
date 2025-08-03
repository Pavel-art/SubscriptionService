package subscriptions

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

type SubscriptionHandler struct {
	logger *zap.Logger
	repo   ISubscriptionRepository
}

func NewSubscriptionHandler(logger *zap.Logger, repo ISubscriptionRepository) *SubscriptionHandler {
	return &SubscriptionHandler{
		logger: logger,
		repo:   repo,
	}
}

func (h *SubscriptionHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		subs := api.Group("/subscriptions")
		{
			subs.POST("", h.Create)
			subs.GET("", h.List)
			subs.GET("/:id", h.Get)
			subs.PUT("/:id", h.Update)
			subs.DELETE("/:id", h.Delete)
		}
	}
}

// CRUD методы
func (h *SubscriptionHandler) Create(c *gin.Context) {
	// Реализация создания
	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

func (h *SubscriptionHandler) Get(c *gin.Context) {
	// Реализация получения
}

func (h *SubscriptionHandler) Update(c *gin.Context) {
	// Реализация обновления
}

func (h *SubscriptionHandler) Delete(c *gin.Context) {
	// Реализация удаления
}

func (h *SubscriptionHandler) List(c *gin.Context) {
	// Реализация получения списка
}
