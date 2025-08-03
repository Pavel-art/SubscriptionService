package subscriptions

import "time"

type CreateSubscriptionRequest struct {
	ServiceName string    `json:"service_name" binding:"required"`
	Price       int       `json:"price" binding:"required,min=1"`
	UserID      string    `json:"user_id" binding:"required"`
	StartDate   time.Time `json:"start_date" binding:"required"`
	EndDate     time.Time `json:"end_date"`
}

type UpdateSubscriptionRequest struct {
	ServiceName *string    `json:"service_name"`
	Price       *int       `json:"price" binding:"omitempty,min=1"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}

type ListSubscriptionsRequest struct {
	UserID      string `form:"user_id"`
	ServiceName string `form:"service_name"`
	Page        int    `form:"page,default=1"`
	Limit       int    `form:"limit,default=10"`
}

type SubscriptionResponse struct {
	ID          string     `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      string     `json:"user_id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type ListResponse struct {
	Data  []SubscriptionResponse `json:"data"`
	Total int                    `json:"total"`
	Page  int                    `json:"page"`
}
