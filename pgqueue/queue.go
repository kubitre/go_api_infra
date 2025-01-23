package pgqueue

import (
	"database/sql"
	"fmt"

	"go.dataddo.com/pgq/x/schema"
)

func InitQueue(db *sql.DB, queueName string) error {
	createQuery := schema.GenerateCreateTableQuery(queueName)
	_, err := db.Exec(createQuery)
	if err != nil {
		return fmt.Errorf("ошибка при инициализации очереди: %w", err)
	}
	return nil
}
