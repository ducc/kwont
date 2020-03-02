package dataservice

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"time"
)

type database struct {
	db *sql.DB
}

// "postgresql://maxroach@localhost:26257/bank?sslmode=disable"
func newDatabase(ctx context.Context, connectionString string) (*database, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	return &database{
		db: db,
	}, nil
}

func (d *database) InsertStrategy(ctx context.Context, entryRules, exitRules []byte, status, name, symbolName, symbolBroker string) (string, error) {
	const statement = `INSERT INTO strategies (entry_rules, exit_rules, status, name, symbol_name, symbol_broker, last_evaluated) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING (strategy_id);`

	res := d.db.QueryRowContext(ctx, statement, entryRules, exitRules, status, name, symbolName, symbolBroker, time.Time{})

	var strategyID string
	if err := res.Scan(&strategyID); err != nil {
		return "", err
	}

	return strategyID, nil
}

func (d *database) InsertSymbolPrice(ctx context.Context, symbolName, symbolBroker string, timestamp time.Time, price int64) error {
	const statement = `INSERT INTO symbol_prices (symbol_name, symbol_broker, timestamp, price) VALUES ($1, $2, $3, $4);`

	if _, err := d.db.ExecContext(ctx, statement, symbolName, symbolBroker, timestamp, price); err != nil {
		return err
	}

	return nil
}
