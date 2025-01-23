package pgqueue

import (
	"context"
	"database/sql"
	"fmt"

	"go.dataddo.com/pgq"
)

type Subscriber struct {
	db *sql.DB
}

// NewSubscriber creates a new Subscriber.
// Accepts a database connection (*sql.DB) and returns a pointer to Subscriber.
func NewSubscriber(db *sql.DB) *Subscriber {
	return &Subscriber{
		db: db,
	}
}

// Subscribe subscribes to a queue and processes messages using the provided handler.
// Accepts a context (context.Context), a queue name (queueName), and a message handler (pgq.MessageHandler).
// Returns an error if something goes wrong.
func (s *Subscriber) Subscribe(ctx context.Context, queueName string, handler pgq.MessageHandler) error {
	if err := InitQueue(s.db, queueName); err != nil {
		return fmt.Errorf("failed to initialize queue: %w", err)
	}

	consumer, err := pgq.NewConsumer(s.db, queueName, handler)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	return consumer.Run(ctx)
}
