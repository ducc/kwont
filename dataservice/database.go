package dataservice

import (
	"context"
	"database/sql"
	"github.com/ducc/kwɒnt/protos"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
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

func (d *database) GetPartialCandlesticks(ctx context.Context, symbolName, symbolBroker string, start, end time.Time) ([]*protos.Candlestick, error) {
	const statement = `SELECT timestamp, current FROM candlesticks WHERE symbol_name = $1 and symbol_broker = $2 ORDER BY timestamp ASC`
	logrus.Debugf("getting partial candlesticks with symbol name %s symbol broker %s start %s end %s", symbolName, symbolBroker, start.String(), end.String())

	iter, err := d.db.QueryContext(ctx, statement, symbolName, symbolBroker)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	output := make([]*protos.Candlestick, 0)

	for iter.Next() {
		var timestamp time.Time
		var current int64

		if err := iter.Scan(&timestamp, &current); err != nil {
			return nil, err
		}

		ts, err := ptypes.TimestampProto(timestamp)
		if err != nil {
			return nil, err
		}

		output = append(output, &protos.Candlestick{
			Timestamp: ts,
			Current:   current,
		})
	}

	return output, nil
}

func (d *database) InsertCandlestick(ctx context.Context, symbolName, symbolBroker string, timestamp time.Time, open, close, high, low, current, spread, buyVolume, sellVolume int64) error {
	const statement = `INSERT INTO candlesticks (symbol_name, symbol_broker, timestamp, open, close, high, low, current, spread, buy_volume, sell_volume) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`

	if _, err := d.db.ExecContext(ctx, statement, symbolName, symbolBroker, timestamp, open, close, high, low, current, spread, buyVolume, sellVolume); err != nil {
		return err
	}

	return nil
}

func (d *database) InsertStrategy(ctx context.Context, entryRules, exitRules []byte, status, name, symbolName, symbolBroker string) (string, error) {
	const statement = `INSERT INTO strategies (entry_rules, exit_rules, status, name, symbol_name, symbol_broker, last_evaluated) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING (strategy_id);`

	row := d.db.QueryRow(statement, entryRules, exitRules, status, name, symbolName, symbolBroker, time.Time{})
	var strategyID string
	if err := row.Scan(&strategyID); err != nil {
		return "", err
	}

	return strategyID, nil
}

func (d *database) UpdateStrategy(ctx context.Context, strategyID string, entryRules, exitRules []byte, status, name, symbolName, symbolBroker string, lastEvaluated time.Time) error {
	const statement = `
		UPDATE strategies 
		SET entry_rules = $1, 
			exit_rules = $2, 
			status = $3, 
			name = $4, 
			symbol_name = $5, 
			symbol_broker = $6,
		    last_evaluated = $7
		WHERE strategy_id = $8`

	if _, err := d.db.ExecContext(ctx, statement, entryRules, exitRules, status, name, symbolName, symbolBroker, lastEvaluated, strategyID); err != nil {
		return err
	}

	return nil
}

func (d *database) ListStrategies(ctx context.Context) ([]*protos.Strategy, error) {
	const statement = `
		SELECT 
			strategy_id,
			entry_rules,
			exit_rules,
			status,
			name,
			symbol_name,
			symbol_broker,
			last_evaluated
		FROM strategies`

	rows, err := d.db.QueryContext(ctx, statement)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logrus.WithError(err).Error("closing rows")
		}
	}()

	strategies := make([]*protos.Strategy, 0)

	for rows.Next() {
		strategy := &protos.Strategy{}

		var entryRules, exitRules []byte
		var status, symbolName, symbolBroker string
		var lastEvaluated time.Time

		if err := rows.Scan(&strategy.Id, &entryRules, &exitRules, &status, &strategy.Name, &symbolName, &symbolBroker, &lastEvaluated); err != nil {
			return nil, err
		}

		var entryRulesSet protos.RuleSet
		if err := proto.Unmarshal(entryRules, &entryRulesSet); err != nil {
			return nil, err
		}

		var exitRulesSet protos.RuleSet
		if err := proto.Unmarshal(exitRules, &exitRulesSet); err != nil {
			return nil, err
		}

		strategy.EntryRules = &entryRulesSet
		strategy.ExitRules = &exitRulesSet

		strategy.Status = protos.Status_Name(protos.Status_Name_value[status])

		strategy.Symbol = &protos.Symbol{
			Name:   protos.Symbol_Name(protos.Symbol_Name_value[symbolName]),
			Broker: protos.Broker_Name(protos.Broker_Name_value[symbolBroker]),
		}

		ts, err := ptypes.TimestampProto(lastEvaluated)
		if err != nil {
			return nil, err
		}
		strategy.LastEvaluated = ts

		strategies = append(strategies, strategy)
	}

	return strategies, nil
}
