package subscriptions

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"time"
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
			subs.GET("/cost", h.CalculateCost)
		}
	}
}

// Create godoc
// @Summary Create subscription
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param subscription body CreateSubscriptionRequest true "Subscription"
// @Success 201 {object} Subscription
// @Failure 400,500 {object} gin.H
// @Router /subscriptions [post]
func (h *SubscriptionHandler) Create(c *gin.Context) {
	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, use MM-YYYY"})
		return
	}

	var endDatePtr *time.Time
	if req.EndDate != "" {
		endDate, err := time.Parse("01-2006", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use MM-YYYY"})
			return
		}
		endDatePtr = &endDate
	}

	sub, err := NewSubscription(req.ServiceName, req.Price, req.UserID, startDate, endDatePtr)
	if err != nil {
		h.logger.Error("invalid subscription", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Create(c.Request.Context(), sub); err != nil {
		h.logger.Error("failed to create subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription"})
		return
	}

	c.JSON(http.StatusCreated, sub)
}

// Get godoc
// @Summary Get subscription by ID
// @Tags Subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} Subscription
// @Failure 404,500 {object} gin.H
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) Get(c *gin.Context) {
	id := c.Param("id")
	sub, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscription"})
		return
	}
	if sub == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}
	c.JSON(http.StatusOK, sub)
}

// Update godoc
// @Summary Update subscription
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param subscription body UpdateSubscriptionRequest true "Subscription"
// @Success 200 {object} Subscription
// @Failure 400,500 {object} gin.H
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, use MM-YYYY"})
		return
	}

	var endDatePtr *time.Time
	if req.EndDate != "" {
		endDate, err := time.Parse("01-2006", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use MM-YYYY"})
			return
		}
		endDatePtr = &endDate
	}

	sub, err := NewSubscription(req.ServiceName, req.Price, req.UserID, startDate, endDatePtr)
	if err != nil {
		h.logger.Error("invalid subscription update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sub.ID = id

	if err := h.repo.Update(c.Request.Context(), sub); err != nil {
		h.logger.Error("failed to update subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// Delete godoc
// @Summary Delete subscription
// @Tags Subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 204
// @Failure 500 {object} gin.H
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete subscription"})
		return
	}
	c.Status(http.StatusNoContent)
}

// List godoc
// @Summary List subscriptions
// @Tags Subscriptions
// @Produce json
// @Param user_id query string false "User ID"
// @Param service_name query string false "Service Name"
// @Param start_date_from query string false "Start Date From MM-YYYY"
// @Param start_date_to query string false "Start Date To MM-YYYY"
// @Success 200 {array} Subscription
// @Failure 400,500 {object} gin.H
// @Router /subscriptions [get]
func (h *SubscriptionHandler) List(c *gin.Context) {
	filters := make(map[string]interface{})

	if serviceName := c.Query("service_name"); serviceName != "" {
		filters["service_name"] = serviceName
	}
	if userID := c.Query("user_id"); userID != "" {
		filters["user_id"] = userID
	}
	if startDateFrom := c.Query("start_date_from"); startDateFrom != "" {
		date, err := time.Parse("01-2006", startDateFrom)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date_from format, use MM-YYYY"})
			return
		}
		filters["start_date_from"] = date
	}
	if startDateTo := c.Query("start_date_to"); startDateTo != "" {
		date, err := time.Parse("01-2006", startDateTo)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date_to format, use MM-YYYY"})
			return
		}
		filters["start_date_to"] = date
	}

	subs, err := h.repo.List(c.Request.Context(), filters)
	if err != nil {
		h.logger.Error("failed to list subscriptions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list subscriptions"})
		return
	}
	c.JSON(http.StatusOK, subs)
}

// CalculateCostResponse is the response for cost calculation
type CalculateCostResponse struct {
	TotalCost int `json:"total_cost"`
}

// CalculateCost godoc
// @Summary Calculate total cost of subscriptions
// @Tags Subscriptions
// @Produce json
// @Param user_id query string false "User ID"
// @Param service_name query string false "Service Name"
// @Param start_date_from query string false "Start Date From MM-YYYY"
// @Param start_date_to query string false "Start Date To MM-YYYY"
// @Success 200 {object} CalculateCostResponse
// @Failure 400,500 {object} gin.H
// @Router /subscriptions/cost [get]
func (h *SubscriptionHandler) CalculateCost(c *gin.Context) {
	filters := make(map[string]interface{})

	if serviceName := c.Query("service_name"); serviceName != "" {
		filters["service_name"] = serviceName
	}
	if userID := c.Query("user_id"); userID != "" {
		filters["user_id"] = userID
	}
	if startDateFrom := c.Query("start_date_from"); startDateFrom != "" {
		date, err := time.Parse("01-2006", startDateFrom)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date_from format, use MM-YYYY"})
			return
		}
		filters["start_date_from"] = date
	}
	if startDateTo := c.Query("start_date_to"); startDateTo != "" {
		date, err := time.Parse("01-2006", startDateTo)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date_to format, use MM-YYYY"})
			return
		}
		filters["start_date_to"] = date
	}

	total, err := h.repo.CalculateMonthlyCost(c.Request.Context(), filters)
	if err != nil {
		h.logger.Error("failed to calculate monthly cost", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate monthly cost"})
		return
	}

	c.JSON(http.StatusOK, CalculateCostResponse{TotalCost: total})
}
