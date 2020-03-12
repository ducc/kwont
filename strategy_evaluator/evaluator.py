import pandas as pd
import protos_pb2
from dataframe import candlesticks_to_dataframe
from indicators.relative_strength_index import evaluate_relative_strength_index
from indicators.relative_strength_index import UnknownConditionError
import traceback

def evaluate(strategy: protos_pb2.Strategy, candlesticks, has_open_position):
    df = candlesticks_to_dataframe(candlesticks)
    # todo re window candlesticks per each rule period

    rules = [] # todo learn how scoping works in python
    if has_open_position:
        rules = strategy.entry_rules.rules
    else:
        rules = strategy.exit_rules.rules

    result = False
    try:
        result = evaluate_rules(rules, df)
    except InvalidRuleError as e:
        return None

    if result:
        action = protos_pb2.EvaluateStrategyResponse.Action()

        if not has_open_position:
            # todo determine which way to open/close position
            action.open_position.price = 999 # todo is price necessary
        else:
            action.close_position.price = 111

        return action
    else:
        return None

def evaluate_rules(rules: protos_pb2.RuleSet, candlesticks):
    # todo allow OR cond instead of just AND
    for rule in rules:
        try:
            if not evaluate_rule(rule, candlesticks):
                return False
        except InvalidRuleError as e:
            raise e

    return True

def evaluate_rule(rule: protos_pb2.Rule, candlesticks):
    indicator = rule.indicator

    series = 0 # todo learn how scoping works in python
    if rule.price_type == protos_pb2.PriceType.OPEN:
        series = candlesticks['open']
    elif rule.price_type == protos_pb2.PriceType.CLOSE:
        series = candlesticks['close']
    else:
        raise InvalidRuleError("unknown price type")

    if indicator.HasField("simple_moving_average"):
        raise Exception("sma not implemented")
    elif indicator.HasField("relative_strength_index"):
        try:
            result = evaluate_relative_strength_index(indicator.relative_strength_index, series)
            return result
        except UnknownConditionError as e:
            raise InvalidRuleError(e)
        except Exception:
            traceback.print_exc()
    else:
        raise InvalidRuleError("unknown indicator")

class InvalidRuleError(Exception):
    pass

