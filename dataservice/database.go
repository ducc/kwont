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

func newDatabase(ctx context.Context, connectionString string) (*database, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &database{
		db: db,
	}, nil
}

/*func (d *database) InsertStrategy(ctx context.Context, entryRules, exitRules []byte, status, name, symbolName, symbolBroker string) (string, error) {
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
*/

func (d *database) InsertCandlestick(ctx context.Context, symbolName, symbolBroker string, timestamp time.Time, open, close, high, low, current, spread, buyVolume, sellVolume int64) error {
	const statement = `INSERT INTO candlesticks (symbol_name, symbol_broker, timestamp, open, close, high, low, current, spread, buy_volume, sell_volume) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`

	if _, err := d.db.ExecContext(ctx, statement, symbolName, symbolBroker, timestamp, open, close, high, low, current, spread, buyVolume, sellVolume); err != nil {
		return err
	}

	return nil
}
