// Package pgqueue provides a generic PostgreSQL queue implementation using the pgq library.
// It allows publishing and subscribing to messages of any type, with automatic queue initialization.
//
// Example Usage:
//
//	// Create a new Subscriber for a specific payload type.
//	sub := pgqueue.NewSubscriber[MyPayload](db)
//
//	// Define a handler function to process messages.
//	handler := func(payload MyPayload) error {
//	    fmt.Printf("Received message: %+v\n", payload)
//	    return nil
//	}
//
//	// Subscribe to a queue.
//	err := sub.Subscribe(context.Background(), "my_queue", handler)
//	if err != nil {
//	    log.Fatal("Failed to start subscriber: ", err)
//	}
package pgqueue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"go.dataddo.com/pgq"
)

type Subscriber[T any] struct {
	db *sql.DB
}

// NewSubscriber creates a new Subscriber for a specific payload type.
// It accepts a database connection (*sql.DB) and returns a pointer to Subscriber[T].
//
// Example:
//
//	sub := pgqueue.NewSubscriber[MyPayload](db)
func NewSubscriber[T any](db *sql.DB) *Subscriber[T] {
	return &Subscriber[T]{
		db: db,
	}
}

// Subscribe subscribes to a queue and processes messages using the provided handler.
// It accepts a context (context.Context), a queue name (queueName), and a message handler (handler).
// The handler function processes messages of type T and returns an error if processing fails.
func (s *Subscriber[T]) Subscribe(ctx context.Context, queueName string, handler func(payload T) error) error {
	if err := InitQueue(s.db, queueName); err != nil {
		return fmt.Errorf("failed to initialize queue: %w", err)
	}

	baseHandler := s.baseHandler(handler)

	consumer, err := pgq.NewConsumer(s.db, queueName, baseHandler)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	return consumer.Run(ctx)
}

func (s *Subscriber[T]) baseHandler(handler func(payload T) error) pgq.MessageHandler {
	return pgq.MessageHandlerFunc(func(ctx context.Context, mi *pgq.MessageIncoming) (bool, error) {
		var payload T
		if err := json.Unmarshal(mi.Payload, &payload); err != nil {
			return false, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		if err := handler(payload); err != nil {
			return false, fmt.Errorf("handler failed: %w", err)
		}

		return true, nil
	})
}
