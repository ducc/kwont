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