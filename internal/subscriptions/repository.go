package subscriptions

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
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

var _ ISubscriptionRepository = (*SubscriptionRepository)(nil)

type SubscriptionRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewSubscriptionRepository(db *pgxpool.Pool, logger *zap.Logger) *SubscriptionRepository {
	return &SubscriptionRepository{db: db, logger: logger}
}

func (s SubscriptionRepository) Create(ctx context.Context, sub *Subscription) error {

	if sub.ID == "" {
		sub.ID = uuid.New().String()
	}

	query := `
        INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id`

	// Проверяем контекст перед выполнением запроса
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error before query: %w", err)
	}

	// Выполняем запрос с учетом возможного NULL для end_date
	var endDate interface{} = nil
	if sub.EndDate != nil {
		endDate = *sub.EndDate
	}

	err := s.db.QueryRow(ctx, query,
		sub.ID,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		endDate,
	).Scan(&sub.ID)

	if err != nil {
		// Логгируем ошибку, если логгер доступен
		if s.logger != nil {
			s.logger.Error("failed to create subscription",
				zap.Error(err),
				zap.String("service", sub.ServiceName),
				zap.String("user", sub.UserID))
		}
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

func (s SubscriptionRepository) GetByID(ctx context.Context, id string) (*Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE id = $1`

	s.logger.Debug("Executing GetByID query",
		zap.String("query", query),
		zap.String("id", id))

	row := s.db.QueryRow(ctx, query, id)
	sub := &Subscription{}

	err := row.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate)
	if err != nil {
		s.logger.Error("Failed to get subscription by ID",
			zap.String("id", id),
			zap.Error(err))

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("subscription not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	s.logger.Debug("Successfully retrieved subscription",
		zap.String("id", sub.ID),
		zap.String("service", sub.ServiceName))

	return sub, nil
}

func (s SubscriptionRepository) Update(ctx context.Context, sub *Subscription) error {
	query := `
        UPDATE subscriptions 
        SET service_name = $1, price = $2, user_id = $3, start_date = $4, end_date = $5
        WHERE id = $6`

	s.logger.Debug("Executing Update query",
		zap.String("query", query),
		zap.Any("subscription", sub))

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
		sub.ID)

	if err != nil {
		s.logger.Error("Failed to update subscription",
			zap.String("id", sub.ID),
			zap.Error(err))
		return fmt.Errorf("database error: %w", err)
	}

	if result.RowsAffected() == 0 {
		s.logger.Warn("No rows affected when updating subscription",
			zap.String("id", sub.ID))
		return fmt.Errorf("subscription not found")
	}

	s.logger.Info("Successfully updated subscription",
		zap.String("id", sub.ID))

	return nil
}

func (s SubscriptionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	s.logger.Debug("Executing Delete query",
		zap.String("query", query),
		zap.String("id", id))

	result, err := s.db.Exec(ctx, query, id)
	if err != nil {
		s.logger.Error("Failed to delete subscription",
			zap.String("id", id),
			zap.Error(err))
		return fmt.Errorf("database error: %w", err)
	}

	if result.RowsAffected() == 0 {
		s.logger.Warn("No subscription found to delete",
			zap.String("id", id))
		return fmt.Errorf("subscription not found")
	}

	s.logger.Info("Successfully deleted subscription",
		zap.String("id", id))

	return nil
}

func (s SubscriptionRepository) List(ctx context.Context, filters map[string]interface{}) ([]*Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions`
	var args []interface{}
	conditions := ""
	i := 1

	for k, v := range filters {
		if conditions == "" {
			conditions = " WHERE "
		} else {
			conditions += " AND "
		}
		conditions += fmt.Sprintf("%s = $%d", k, i)
		args = append(args, v)
		i++
	}

	fullQuery := query + conditions

	s.logger.Debug("Executing List query",
		zap.String("query", fullQuery),
		zap.Any("filters", filters))

	rows, err := s.db.Query(ctx, fullQuery, args...)
	if err != nil {
		s.logger.Error("Failed to list subscriptions",
			zap.Any("filters", filters),
			zap.Error(err))
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var subs []*Subscription
	for rows.Next() {
		var sub Subscription
		err := rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate)
		if err != nil {
			s.logger.Error("Failed to scan subscription row",
				zap.Error(err))
			return nil, fmt.Errorf("database scan error: %w", err)
		}
		subs = append(subs, &sub)
	}

	s.logger.Debug("Successfully listed subscriptions",
		zap.Int("count", len(subs)))

	return subs, nil
}

func (s SubscriptionRepository) CalculateMonthlyCost(ctx context.Context, filters map[string]interface{}) (int, error) {
	query := `SELECT SUM(price) FROM subscriptions`
	var args []interface{}
	conditions := ""
	i := 1

	for k, v := range filters {
		if conditions == "" {
			conditions = " WHERE "
		} else {
			conditions += " AND "
		}
		conditions += fmt.Sprintf("%s = $%d", k, i)
		args = append(args, v)
		i++
	}

	fullQuery := query + conditions

	s.logger.Debug("Executing CalculateMonthlyCost query",
		zap.String("query", fullQuery),
		zap.Any("filters", filters))

	var sum int
	err := s.db.QueryRow(ctx, fullQuery, args...).Scan(&sum)
	if err != nil {
		s.logger.Error("Failed to calculate monthly cost",
			zap.Any("filters", filters),
			zap.Error(err))
		return 0, fmt.Errorf("database error: %w", err)
	}

	s.logger.Info("Calculated monthly cost",
		zap.Int("total", sum),
		zap.Any("filters", filters))

	return sum, nil
}
