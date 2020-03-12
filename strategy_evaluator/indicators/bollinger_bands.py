import protos_pb2

from indicators.errors import UnknownConditionError

class BollingerBands():
    def __init__(self, values, window_periods, deviation):
        self.values = values
        self.ma = values.rolling(window = window_periods).mean()
        std = values.rolling(window = window_periods).std()

        self.upper = self.ma + (std * deviation)
        self.lower = self.ma - (std * deviation)

    def plot(self):
        self.values.plot()
        self.upper.plot()
        self.lower.plot()
        self.ma.plot()

    # tuple of (upper, lower, ma)
    def current_values(self):
        return self.upper.tail(1)[0], self.lower.tail(1)[0], self.ma.tail(1)[0]

    def is_above_upper_band(self):
        upper, lower, ma = self.current_values()
        return self.values.tail(1)[0] > upper

    def is_below_upper_band(self):
        upper, lower, ma = self.current_values()
        return self.values.tail(1)[0] < upper

    def is_above_lower_band(self):
        upper, lower, ma = self.current_values()
        return self.values.tail(1)[0] > lower

    def is_below_lower_band(self):
        upper, lower, ma = self.current_values()
        return self.values.tail(1)[0] < lower

    def is_above_moving_average(self):
        upper, lower, ma = self.current_values()
        return self.values.tail(1)[0] > ma

    def is_below_moving_average(self):
        upper, lower, ma = self.current_values()
        return self.values.tail(1)[0] < ma

def evaluate_bollinger_bands(indicator: protos_pb2.Rule.Indicator.MACD, series):
    bb = BollingerBands(series, window_periods=indicator.period, deviation=indicator.deviation)

    if indicator.condition == protos_pb2.Rule.Indicator.BollingerBands.Condition.PRICE_ABOVE_UPPER_BAND:
        return bb.is_above_upper_band()
    elif indicator.condition == protos_pb2.Rule.Indicator.BollingerBands.Condition.PRICE_BELOW_UPPER_BAND:
        return bb.is_below_upper_band()
    elif indicator.condition == protos_pb2.Rule.Indicator.BollingerBands.Condition.PRICE_ABOVE_LOWER_BAND:
        return bb.is_above_lower_band()
    elif indicator.condition == protos_pb2.Rule.Indicator.BollingerBands.Condition.PRICE_BELOW_LOWER_BAND:
        return bb.is_below_lower_band()
    elif indicator.condition == protos_pb2.Rule.Indicator.BollingerBands.Condition.PRICE_ABOVE_MA:
        return bb.is_above_moving_average()
    elif indicator.condition == protos_pb2.Rule.Indicator.BollingerBands.Condition.PRICE_BELOW_MA:
        return bb.is_below_moving_average()
    else:
        raise UnknownConditionError("unknown bollinger bands condition")
