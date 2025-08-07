package subscriptions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type ISubscriptionRepository interface {
	Create(ctx context.Context, sub *Subscription) error
	GetByID(ctx context.Context, id string) (*Subscription, error)
	Update(ctx context.Context, sub *Subscription) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters map[string]interface{}) ([]*Subscription, error)
	CalculateMonthlyCost(ctx context.Context, filters map[string]interface{}) (int, error)
}

type SubscriptionRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewSubscriptionRepository(db *pgxpool.Pool, logger *zap.Logger) *SubscriptionRepository {
	return &SubscriptionRepository{db: db, logger: logger}
}

func (s *SubscriptionRepository) Create(ctx context.Context, sub *Subscription) error {
	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	var endDate any = nil
	if sub.EndDate != nil {
		endDate = *sub.EndDate
	}

	err := s.db.QueryRow(ctx, query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		endDate,
	).Scan(&sub.ID)

	if err != nil {
		s.logger.Error("failed to create subscription",
			zap.Error(err),
			zap.String("service", sub.ServiceName),
			zap.String("user", sub.UserID))
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

func (s *SubscriptionRepository) GetByID(ctx context.Context, id string) (*Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date
		FROM subscriptions 
		WHERE id = $1`

	sub := &Subscription{}
	var endDate *time.Time

	err := s.db.QueryRow(ctx, query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&endDate,
	)

	sub.EndDate = endDate

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("subscription not found")
		}
		s.logger.Error("failed to get subscription",
			zap.Error(err),
			zap.String("id", id))
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return sub, nil
}

func (s *SubscriptionRepository) Update(ctx context.Context, sub *Subscription) error {
	query := `
		UPDATE subscriptions 
		SET service_name = $1, price = $2, user_id = $3, 
			start_date = $4, end_date = $5
		WHERE id = $6`

	var endDate interface{} = nil
	if sub.EndDate != nil {
		endDate = *sub.EndDate
	}

	result, err := s.db.Exec(ctx, query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		endDate,
		sub.ID,
	)

	if err != nil {
		s.logger.Error("failed to update subscription",
			zap.Error(err),
			zap.String("id", sub.ID))
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("subscription not found")
	}

	return nil
}

func (s *SubscriptionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	result, err := s.db.Exec(ctx, query, id)
	if err != nil {
		s.logger.Error("failed to delete subscription",
			zap.Error(err),
			zap.String("id", id))
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("subscription not found")
	}

	return nil
}

func (s *SubscriptionRepository) List(ctx context.Context, filters map[string]interface{}) ([]*Subscription, error) {
	baseQuery := `
		SELECT id, service_name, price, user_id, start_date, end_date
		FROM subscriptions`

	var conditions []string
	var args []interface{}

	for field, value := range filters {
		conditions = append(conditions, fmt.Sprintf("%s = $%d", field, len(args)+1))
		args = append(args, value)
	}

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	rows, err := s.db.Query(ctx, baseQuery, args...)
	if err != nil {
		s.logger.Error("failed to list subscriptions",
			zap.Error(err))
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []*Subscription
	for rows.Next() {
		var sub Subscription
		var endDate *time.Time

		if err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&endDate,
		); err != nil {
			s.logger.Error("failed to scan subscription",
				zap.Error(err))
			continue
		}

		sub.EndDate = endDate
		subs = append(subs, &sub)
	}

	return subs, nil
}

func (s *SubscriptionRepository) CalculateMonthlyCost(ctx context.Context, filters map[string]interface{}) (int, error) {
	baseQuery := `SELECT COALESCE(SUM(price), 0) FROM subscriptions`

	var conditions []string
	var args []any

	for field, value := range filters {
		conditions = append(conditions, fmt.Sprintf("%s = $%d", field, len(args)+1))
		args = append(args, value)
	}

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := s.db.QueryRow(ctx, baseQuery, args...).Scan(&total)
	if err != nil {
		s.logger.Error("failed to calculate monthly cost",
			zap.Error(err))
		return 0, fmt.Errorf("failed to calculate monthly cost: %w", err)
	}

	return total, nil
}
