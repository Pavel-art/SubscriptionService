package subscriptions

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidServiceName = errors.New("service name must be between 2 and 100 characters")
	ErrInvalidPrice       = errors.New("price must be positive")
	ErrInvalidUserID      = errors.New("invalid user ID format")
	ErrInvalidDateRange   = errors.New("end date must be after start date")
)

type Subscription struct {
	ID          string     `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      string     `json:"user_id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func NewSubscription(
	serviceName string,
	price int,
	userID string,
	startDate time.Time,
	endDate *time.Time,
) (*Subscription, error) {
	sub := &Subscription{
		ServiceName: strings.TrimSpace(serviceName),
		Price:       price,
		UserID:      strings.TrimSpace(userID),
		StartDate:   startDate,
		EndDate:     endDate,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := sub.Validate(); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *Subscription) Validate() error {

	s.ServiceName = strings.TrimSpace(s.ServiceName)
	if len(s.ServiceName) < 2 || len(s.ServiceName) > 100 {
		return ErrInvalidServiceName
	}

	if s.Price <= 0 {
		return ErrInvalidPrice
	}

	s.UserID = strings.TrimSpace(s.UserID)
	if len(s.UserID) != 36 {
		return ErrInvalidUserID
	}

	if s.EndDate != nil && s.EndDate.Before(s.StartDate) {
		return ErrInvalidDateRange
	}

	return nil
}
