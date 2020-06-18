package dataservice

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ducc/kwɒnt/protos"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"strings"
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
	db.SetMaxIdleConns(10)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &database{
		db: db,
	}, nil
}

func (d *database) GetCandlesticks(ctx context.Context, window protos.CandlestickWindow_Name, broker protos.Broker_Name, symbol protos.Symbol_Name, start, end time.Time) ([]*protos.Candlestick, error) {
	table := getCandlestickTableFromWindow(window)
	if table == "" {
		return nil, errors.New("unsupported window")
	}

	statement := fmt.Sprintf("SELECT timestamp, open_price, close_price, high_price, low_price, buy_volume, sell_volume FROM %s WHERE timestamp >= $1 AND timestamp <= $2 and broker = $3 and symbol = $4 ORDER BY timestamp ASC;", table)

	iter, err := d.db.QueryContext(ctx, statement, start, end, broker.String(), symbol.String())
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	output := make([]*protos.Candlestick, 0)

	for iter.Next() {
		var timestamp time.Time
		candlestick := &protos.Candlestick{
			Broker: broker,
			Symbol: symbol,
		}

		if err := iter.Scan(&timestamp, &candlestick.OpenPrice, &candlestick.ClosePrice, &candlestick.HighPrice, &candlestick.LowPrice, &candlestick.BuyVolume, &candlestick.SellVolume); err != nil {
			return nil, err
		}

		candlestick.Timestamp, err = ptypes.TimestampProto(timestamp)
		if err != nil {
			return nil, err
		}

		output = append(output, candlestick)
	}

	return output, nil
}

func (d *database) InsertTick(ctx context.Context, timestamp time.Time, broker, symbol string, price, spread, buyVolume, sellVolume float64) error {
	const statement = `INSERT INTO ticks (timestamp, broker, symbol, price, spread, buy_volume, sell_volume) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING;`

	if _, err := d.db.ExecContext(ctx, statement, timestamp, broker, symbol, price, spread, buyVolume, sellVolume); err != nil {
		return err
	}

	return nil
}

func truncateTimestampForWindow(window protos.CandlestickWindow_Name, timestamp time.Time) time.Time {
	switch window {
	case protos.CandlestickWindow_ONE_MINUTE:
		return timestamp.Truncate(time.Minute)
	case protos.CandlestickWindow_ONE_HOUR:
		return timestamp.Truncate(time.Hour)
	case protos.CandlestickWindow_ONE_DAY:
		return timestamp.Truncate(time.Hour * 24)
	default:
		return time.Time{}
	}
}

func getCandlestickTableFromWindow(window protos.CandlestickWindow_Name) string {
	switch window {
	case protos.CandlestickWindow_ONE_MINUTE:
		return "candlesticks_1m"
	case protos.CandlestickWindow_ONE_HOUR:
		return "candlesticks_1h"
	case protos.CandlestickWindow_ONE_DAY:
		return "candlesticks_1d"
	default:
		return ""
	}
}

func (d *database) InsertOrUpdateCandlestick(ctx context.Context, window protos.CandlestickWindow_Name, timestamp time.Time, broker, symbol string, price, spread, buyVolume, sellVolume float64) error {
	windowedTime := truncateTimestampForWindow(window, timestamp)
	table := getCandlestickTableFromWindow(window)

	if windowedTime.IsZero() || table == "" {
		return errors.New("unsupported window")
	}

	// todo use spread
	const query = `
insert into TABLENAME values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
on conflict (timestamp, broker, symbol) do update set 
updated = case when TABLENAME.updated < excluded.updated then excluded.updated else TABLENAME.updated end,
open_price = case when TABLENAME.updated > excluded.updated then excluded.open_price else TABLENAME.open_price end, 
close_price = case when TABLENAME.updated < excluded.updated then excluded.close_price else TABLENAME.close_price end, 
high_price = case when TABLENAME.high_price < excluded.high_price then excluded.high_price else TABLENAME.high_price end, 
low_price = case when TABLENAME.low_price > excluded.low_price then excluded.low_price else TABLENAME.low_price end, 
buy_volume = TABLENAME.buy_volume + excluded.buy_volume, 
sell_volume = TABLENAME.sell_volume + excluded.sell_volume;
	`
	statement := strings.ReplaceAll(query, "TABLENAME", table) // todo BAD DOG BAD DOG BAD DOG BAD DOG BAD DOG

	if _, err := d.db.ExecContext(ctx, statement, windowedTime, broker, symbol, timestamp, price, price, price, price, buyVolume, sellVolume); err != nil {
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

func (d *database) InsertOrder(ctx context.Context, broker, symbol, direction string, price, volume float64, timestamp time.Time) (string, error) {
	const stmt = `INSERT INTO orders (broker, symbol, direction, price, volume, timestamp) VALUES ($1, $2, $3, $4, $5, $6) RETURNING order_id;`
	row := d.db.QueryRowContext(ctx, stmt, broker, symbol, direction, price, volume, timestamp)

	var id string
	if err := row.Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

func (d *database) SelectOrder(ctx context.Context, orderID string) (*protos.Order, error) {
	const stmt = `SELECT broker, symbol, direction, price, volume, timestamp FROM orders WHERE order_id = $1;`
	row := d.db.QueryRowContext(ctx, stmt, orderID)

	var brokerName, symbolName, directionName string
	var timestamp time.Time
	order := &protos.Order{}

	err := row.Scan(&brokerName, &symbolName, &directionName, &order.Price, &order.Volume, &timestamp)
	if err != nil {
		return nil, err
	}

	order.Broker = protos.Broker_Name(protos.Broker_Name_value[brokerName])
	order.Symbol = protos.Symbol_Name(protos.Symbol_Name_value[symbolName])
	order.Direction = protos.Direction_Name(protos.Direction_Name_value[directionName])
	order.Timestamp, err = ptypes.TimestampProto(timestamp)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (d *database) InsertXTBTrade(ctx context.Context, timestamp time.Time, sessionID string, order int64, closePrice float64, closeTime time.Time, closed bool, cmd, comment string, commission float64, customComment string, digits int64, expiration time.Time, marginRate float64, offset int64, openPrice float64, openTime time.Time, order2, position int64, profit, stopLoss float64, state string, storage float64, symbol string, takeProfit float64, tradeType string, volume float64) error {
	const stmt = `
INSERT INTO xtb_trades (session_id, "order", timestamp, close_price, close_time, closed, cmd, 
						comment, commission, custom_comment, digits, expiration, margin_rate, 
						"offset", open_price, open_time, order_2, position, profit, stop_loss, 
						state, storage, symbol, take_profit, type, volume)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 
		$11, $12, $13, $14, $15, $16, $17, $18, $19, 
		$20, $21, $22, $23, $24, $25, $26);
`

	if _, err := d.db.ExecContext(ctx, stmt, sessionID, order, timestamp, closePrice, closeTime, closed, cmd,
		comment, commission, customComment, digits, expiration, marginRate,
		offset, openPrice, openTime, order2, position, profit, stopLoss,
		state, storage, symbol, takeProfit, tradeType, volume); err != nil {
		return err
	}

	return nil
}

func (d *database) InsertXTBTradeStatus(ctx context.Context, timestamp time.Time, sessionID string, order int64, customComment, message string, price float64, requestStatus string) error {
	const stmt = `
INSERT INTO xtb_trade_status (session_id, "order", timestamp, custom_comment, message, price, request_status)
VALUES ($1, $2, $3, $4, $5, $6, $7);
`

	if _, err := d.db.ExecContext(ctx, stmt, sessionID, order, timestamp, customComment, message, price, requestStatus); err != nil {
		return err
	}

	return nil
}

func (d *database) SelectXTBTrades(ctx context.Context, orderID string, sessionID string) ([]*protos.XTBTrade, error) {
	const stmt = `SELECT "order", timestamp, closePrice, closeTime, closed, cmd, comment, commission, digits, 
                         expiration, marginRate, "offset", openPrice, openTime, order2, position, profit, stopLoss,
                         state, storage, symbol, takeProfit, tradeType, volume
                  WHERE custom_comment = $1 AND session_id = $2 
				  ORDER BY timestamp asc`

	rows, err := d.db.QueryContext(ctx, stmt, orderID, sessionID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logrus.WithError(err).Error("closing query rows")
		}
	}()

	trades := make([]*protos.XTBTrade, 0)

	for rows.Next() {
		trade := &protos.XTBTrade{}
		var (
			timestamp, closeTime, expiration, openTime time.Time
			symbol                                     string
		)

		err := rows.Scan(&trade.Order, &timestamp, &trade.ClosePrice, &closeTime, &trade.Closed, &trade.Cmd, &trade.Comment,
			&trade.Commission, &trade.Digits, &expiration, &trade.MarginRate, &trade.Offset, &trade.OpenPrice, openTime,
			&trade.Order2, &trade.Position, &trade.Profit, &trade.StopLoss, &trade.State, &trade.Storage, symbol,
			&trade.TakeProfit, &trade.Type, &trade.Volume)
		if err != nil {
			return nil, err
		}

		trade.Timestamp, err = ptypes.TimestampProto(timestamp)
		if err != nil {
			return nil, err
		}

		trade.CloseTime, err = ptypes.TimestampProto(closeTime)
		if err != nil {
			return nil, err
		}

		trade.Expiration, err = ptypes.TimestampProto(expiration)
		if err != nil {
			return nil, err
		}

		trade.OpenTime, err = ptypes.TimestampProto(openTime)
		if err != nil {
			return nil, err
		}

		trade.Symbol = protos.Symbol_Name(protos.Symbol_Name_value[symbol])

		trades = append(trades, trade)
	}

	return trades, nil
}
