import protos_pb2

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

