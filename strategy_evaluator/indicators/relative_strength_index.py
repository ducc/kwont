import protos_pb2

class RelativeStrengthIndex():
    def __init__(self, values, window_periods, overbought, oversold):
        self.overbought = overbought
        self.oversold = oversold

        delta = values.diff()
        delta = delta[1:]

        up, down = delta.copy(), delta.copy()
        up[up < 0] = 0
        down[down > 0] = 0

        roll_up = up.rolling(window_periods).mean()
        roll_down = down.abs().rolling(window_periods).mean()

        relative_strength = roll_up / roll_down
        self.rsi = 100.0 - (100.0 / (1.0 + relative_strength))

    def plot(self):
        self.rsi.plot()

    def current_value(self):
        return self.rsi.tail(1)[0]

    def is_overbought(self):
        return self.current_value() >= self.overbought

    def is_oversold(self):
        return self.current_value() >= self.overbought

def evaluate_relative_strength_index(indicator: protos_pb2.Rule.Indicator.RelativeStrengthIndex, series):
    rsi = RelativeStrengthIndex(series, window_periods=indicator.period, overbought=indicator.over_bought, oversold=indicator.over_sold)

    if indicator.condition == protos_pb2.Rule.Indicator.RelativeStrengthIndex.Condition.ABOVE_OVER_BOUGHT_LINE:
        return rsi.is_overbought()
    elif indicator.condition == protos_pb2.Rule.Indicator.RelativeStrengthIndex.Condition.BELOW_OVER_BOUGHT_LINE:
        return not rsi.is_overbought()
    elif indicator.condition == protos_pb2.Rule.Indicator.RelativeStrengthIndex.Condition.ABOVE_UNDER_SOLD_LINE:
        return not rsi.is_oversold()
    elif indicator.condition == protos_pb2.Rule.Indicator.RelativeStrengthIndex.Condition.BELOW_UNDER_SOLD_LINE:
        return rsi.is_oversold()
    else:
        raise Exception("unknown rsi condition")