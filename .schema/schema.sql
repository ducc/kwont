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
CREATE TABLE IF NOT EXISTS symbol_candlesticks (
    symbol_name STRING NOT NULL,
    symbol_broker STRING NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    open_timestamp TIMESTAMPTZ NOT NULL,
    close_timestamp TIMESTAMPTZ NOT NULL,
    open INT NOT NULL,
    close INT NOT NULL,
    high INT NOT NULL,
    low INT NOT NULL,
    current INT NOT NULL
);