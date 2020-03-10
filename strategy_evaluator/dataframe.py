import pandas as pd

def candlesticks_to_dataframe(candlesticks):
    data = []

    for candlestick in candlesticks:
        ts = candlestick.timestamp.ToJsonString()
        data.append({
            "time": pd.to_datetime(ts),
            "low":   float(candlestick.low)  / 1000000,
            "high":  float(candlestick.high) / 1000000,
            "open":  float(candlestick.open) / 1000000,
            "close": float(candlestick.close) / 1000000,
        })

    return pd.DataFrame(data).set_index('time')
