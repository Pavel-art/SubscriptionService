package subscriptions

import (
	"errors"
	"fmt"
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

func NewSubscription(serviceName string, price int, userID string, startDate time.Time, endDate *time.Time) (*Subscription, error) {
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
		return nil, fmt.Errorf("invalid subscription: %w", err)
	}

	return sub, nil
}

func (s *Subscription) Validate() error {
	if len(s.ServiceName) < 2 || len(s.ServiceName) > 100 {
		return ErrInvalidServiceName
	}

	if s.Price <= 0 {
		return ErrInvalidPrice
	}

	if len(s.UserID) != 36 { // UUID v4 length
		return ErrInvalidUserID
	}

	if s.EndDate != nil && s.EndDate.Before(s.StartDate) {
		return ErrInvalidDateRange
	}

	return nil
}

// IsActive проверяет активна ли подписка на указанную дату
func (s *Subscription) IsActive(at time.Time) bool {
	if at.Before(s.StartDate) {
		return false
	}

	if s.EndDate == nil {
		return true
	}

	return at.Before(*s.EndDate)
}

// CalculateCostForPeriod рассчитывает стоимость за период
func (s *Subscription) CalculateCostForPeriod(from, to time.Time) int {
	if !s.IsActiveDuringPeriod(from, to) {
		return 0
	}

	months := monthsBetween(
		maxTime(from, s.StartDate),
		minTime(to, s.getEndDateOrMax()),
	)

	return s.Price * months
}

// IsActiveDuringPeriod проверяет активна ли подписка в течение всего периода
func (s *Subscription) IsActiveDuringPeriod(from, to time.Time) bool {
	if to.Before(from) {
		return false
	}

	return !from.Before(s.StartDate) &&
		(s.EndDate == nil || !to.After(*s.EndDate))
}

func (s *Subscription) getEndDateOrMax() time.Time {
	if s.EndDate == nil {
		return time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
	}
	return *s.EndDate
}

func monthsBetween(from, to time.Time) int {
	months := 0
	for from.Before(to) {
		from = from.AddDate(0, 1, 0)
		months++
	}
	return months
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
