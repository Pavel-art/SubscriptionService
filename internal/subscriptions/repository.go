package subscriptions

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
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
	db *pgxpool.Pool
}

func NewSubscriptionRepository(db *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (s SubscriptionRepository) Create(ctx context.Context, sub *Subscription) error {
	//TODO implement me
	panic("implement me")
}

func (s SubscriptionRepository) GetByID(ctx context.Context, id string) (*Subscription, error) {
	//TODO implement me
	panic("implement me")
}

func (s SubscriptionRepository) Update(ctx context.Context, sub *Subscription) error {
	//TODO implement me
	panic("implement me")
}

func (s SubscriptionRepository) Delete(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}

func (s SubscriptionRepository) List(ctx context.Context, filters map[string]interface{}) ([]*Subscription, error) {
	//TODO implement me
	panic("implement me")
}

func (s SubscriptionRepository) CalculateMonthlyCost(ctx context.Context, filters map[string]interface{}) (int, error) {
	//TODO implement me
	panic("implement me")
}
