import protos_pb2

from indicators.errors import UnknownConditionError


class MACD():
    def __init__(self, values, short_term, long_term, macd_length):
        ema_short_term = close.ewm(span=short_term).mean()
        ema_long_term = close.ewm(span=long_term).mean()

        self.macd = ema12 - ema26
        self.signal = self.macd.ewm(span=macd_length).mean()

    def plot(self):
        self.macd.plot()
        self.signal.plot()

    def values(self):
        return self.macd.tail(1)[0], self.signal.tail(1)[0]

    def is_signal_above_macd(self):
        macd, signal = self.values()
        return signal > macd

    def is_signal_below_macd(self):
        macd, signal = self.values()
        return signal < macd

def evaluate_macd(indicator: protos_pb2.Rule.Indicator.MACD, series):
    macd = MACD(series, short_term=indicator.short_term, long_term=indicator.long_term, macd_length=indicator.length)

    if indicator.condition == protos_pb2.Rule.Indicator.MACD.Condition.SIGNAL_LINE_ABOVE_MACD:
        return macd.is_signal_above_macd()
    elif indicator.condition == protos_pb2.Rule.Indicator.MACD.Condition.SIGNAL_LINE_BELOW_MACD:
        return macd.is_signal_below_macd()
    else:
        raise UnknownConditionError("unknown macd condition")
