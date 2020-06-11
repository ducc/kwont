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

func (d *database) InsertTick(ctx context.Context, timestamp time.Time, broker, symbol string, price, spread, buyVolume, sellVolume float64) error {
	const statement = `INSERT INTO ticks (timestamp, broker, symbol, price, spread, buy_volume, sell_volume) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING;`

	if _, err := d.db.ExecContext(ctx, statement, timestamp, broker, symbol, price, spread, buyVolume, sellVolume); err != nil {
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

func (d *database) InsertUser(ctx context.Context, name string) (string, error) {
	const stmt = `INSERT INTO users (name) VALUES ($1) RETURNING user_id;`
	row := d.db.QueryRowContext(ctx, stmt, name)

	var id string
	if err := row.Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

func (d *database) GetUser(ctx context.Context, userID string) (*protos.User, error) {
	const stmt = `SELECT u.name                      as name,
						 COALESCE(b.broker_name, '') as broker_name, 
						 COALESCE(b.username, '')    as username, 
                         COALESCE(b.password, '')    as password, 
					     COALESCE(b.session_id, '')  as session_id
				  FROM users u
                  LEFT JOIN broker_connections b ON b.user_id = u.user_id
				  WHERE u.user_id = $1`

	rows, err := d.db.QueryContext(ctx, stmt, userID)
	if err != nil {
		return nil, err
	}

	user := &protos.User{
		Id:                userID,
		BrokerConnections: make([]*protos.User_BrokerConnection, 0),
	}

	for rows.Next() {
		var name, brokerName, username, password, sessionID string
		if err := rows.Scan(&name, &brokerName, &username, &password, &sessionID); err != nil {
			return nil, err
		}

		user.Name = name
		if brokerName != "" {
			user.BrokerConnections = append(user.BrokerConnections, &protos.User_BrokerConnection{
				Broker:    protos.Broker_Name(protos.Broker_Name_value[brokerName]),
				Username:  username,
				Password:  password,
				SessionId: sessionID,
			})
		}
	}

	return user, nil
}

func (d *database) ListUsers(ctx context.Context) ([]*protos.User, error) {
	const stmt = `SELECT u.user_id                   as user_id, 
						 u.name                      as name, 
						 COALESCE(b.broker_name, '') as broker_name, 
						 COALESCE(b.username, '')    as username, 
                         COALESCE(b.password, '')    as password, 
					     COALESCE(b.session_id, '')  as session_id
				  FROM users u
				  LEFT JOIN broker_connections b ON b.user_id = u.user_id
                  ORDER BY b.user_id ASC` // this algorithm depends on the user ids being in order

	rows, err := d.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}

	users := make([]*protos.User, 0)

	var lastUser *protos.User

	for rows.Next() {
		var userID, name, brokerName, username, password, sessionID string
		if err := rows.Scan(&userID, &name, &brokerName, &username, &password, &sessionID); err != nil {
			return nil, err
		}

		if lastUser == nil {
			lastUser = &protos.User{
				Id:                userID,
				Name:              name,
				BrokerConnections: make([]*protos.User_BrokerConnection, 0),
			}
		} else if lastUser.Id != userID {
			users = append(users, lastUser)
			lastUser = &protos.User{
				Id:                userID,
				Name:              name,
				BrokerConnections: make([]*protos.User_BrokerConnection, 0),
			}
		}

		if brokerName != "" {
			lastUser.BrokerConnections = append(lastUser.BrokerConnections, &protos.User_BrokerConnection{
				Broker:    protos.Broker_Name(protos.Broker_Name_value[brokerName]),
				Username:  username,
				Password:  password,
				SessionId: sessionID,
			})
		}
	}

	if lastUser != nil {
		users = append(users, lastUser)
	}

	return users, nil
}

func (d *database) UpdateUser(ctx context.Context, userID, name string) error {
	const stmt = `UPDATE users SET name = $1 WHERE user_id = $2`
	if _, err := d.db.ExecContext(ctx, stmt, name, userID); err != nil {
		return err
	}

	return nil
}

func (d *database) InsertBrokerConnections(ctx context.Context, userID, brokerName, username, password string) error {
	const stmt = `INSERT INTO broker_connections (user_id, broker_name, username, password, session_id) VALUES ($1, $2, $3, $4, $5);`
	if _, err := d.db.ExecContext(ctx, stmt, userID, brokerName, username, password, ""); err != nil {
		return err
	}

	return nil
}

func (d *database) UpdateBrokerConnection(ctx context.Context, userID, brokerName, username, password, sessionID string) error {
	const stmt = `UPDATE broker_connections
				  SET username = $1,
					  password = $2,
					  session_id = $3
                  WHERE user_id = $4
					AND broker_name = $5;`
	if _, err := d.db.ExecContext(ctx, stmt, username, password, sessionID, userID, brokerName); err != nil {
		return err
	}

	return nil
}
