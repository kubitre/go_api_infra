package pgqueue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.dataddo.com/pgq"
)

type Publisher[T any] struct {
	publisher pgq.Publisher
	db        *sql.DB
}

// NewPublisher creates a new Publisher with support for generics.
// Accepts a database connection (*sql.DB) and returns a pointer to Publisher[T].
func NewPublisher[T any](db *sql.DB) *Publisher[T] {
	return &Publisher[T]{
		publisher: pgq.NewPublisher(db),
		db:        db,
	}
}

// Publish publishes a message to the queue.
// Accepts a context (context.Context), a queue name (queueName), and a payload of any type (T).
// Returns an error if something goes wrong.
func (p *Publisher[T]) Publish(ctx context.Context, queueName string, payload T) ([]uuid.UUID, error) {
	if err := InitQueue(p.db, queueName); err != nil {
		return nil, fmt.Errorf("ошибка при инициализации очереди: %w", err)
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("ошибка при сериализации payload: %w", err)
	}

	metadata := pgq.Metadata{
		"version": "1.0",
	}

	msg := &pgq.MessageOutgoing{
		Payload:  payloadBytes,
		Metadata: metadata,
	}

	return p.publisher.Publish(ctx, queueName, msg)
}
