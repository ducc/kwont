CREATE TABLE IF NOT EXISTS strategies (
    strategy_id UUID NOT NULL DEFAULT gen_random_uuid(),
    entry_rules BYTES NOT NULL, -- encoded protobuf bytes
    exit_rules BYTES NOT NULL,
    status STRING NOT NULL,
    name STRING NOT NULL,
    symbol_name STRING NOT NULL,
    symbol_broker STRING NOT NULL,
    last_evaluated TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS positions (
    position_id UUID NOT NULL DEFAULT gen_random_uuid(),
    strategy_id UUID NOT NULL,
    direction STRING NOT NULL,
    open_price INT NOT NULL,
    close_price INT NOT NULL,
    open_time TIMESTAMPTZ NOT NULL,
    close_time TIMESTAMPTZ NOT NULL
) INTERLEAVE IN PARENT (strategies);

CREATE TABLE IF NOT EXISTS symbol_prices (
    symbol_name STRING NOT NULL,
    symbol_broker STRING NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    price INT NOT NULL
);

-- todo can this be auto populated on insert to symbol_prices?
CREATE TABLE IF NOT EXISTS candlesticks (
    candlestick_id UUID NOT NULL DEFAULT gen_random_uuid(),
    symbol_name STRING NOT NULL,
    symbol_broker STRING NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    open INT NOT NULL,
    close INT NOT NULL,
    high INT NOT NULL,
    low INT NOT NULL,
    current INT NOT NULL,
    spread INT NOT NULL,
    buy_volume INT NOT NULL,
    sell_volume INT NOT NULL
);

CREATE INDEX IF NOT EXISTS candlesticks_symbol_broker_name_idx ON candlesticks (
    symbol_broker ASC,
    symbol_name ASC,
    timestamp ASC
);

CREATE TABLE IF NOT EXISTS users (
    user_id UUID NOT NULL DEFAULT gen_random_uuid(),
    name STRING NOT NULL
);

CREATE TABLE IF NOT EXISTS broker_connections (
    user_id UUID NOT NULL,
    broker_name STRING NOT NULL,
    username STRING NOT NULL,
    password STRING NOT NULL,
    session_id STRING NOT NULL
);








CREATE TABLE IF NOT EXISTS ticks (
    timestamp timestamp,
    broker text,
    symbol text,
    price decimal not null,
    spread decimal not null,
    buy_volume decimal not null,
    sell_volume decimal not null,
    PRIMARY KEY (timestamp, broker, symbol)
);

CREATE TABLE IF NOT EXISTS candlesticks_1m (
    timestamp timestamp,
    broker text,
    symbol text,
    open_price decimal not null,
    close_price decimal not null,
    high_price decimal not null,
    low_price decimal not null,
    buy_volume decimal not null,
    sell_volume decimal not null,
    PRIMARY KEY (timestamp, broker, symbol)
    -- todo min_spread, max_spread, avg_spread
);



