from concurrent import futures
import time
import math
import logging
import grpc
import pandas as pd

import protos_pb2
import protos_pb2_grpc

from evaluator import evaluate
import traceback

class StrategyEvaluatorServer(protos_pb2_grpc.StrategyEvaluatorServicer):
    def Evaluate(self, request: protos_pb2.EvaulateStrategyRequest, context):
        response = protos_pb2.EvaluateStrategyResponse()

        try:
            action = evaluate(request.strategy, request.candlesticks, request.has_open_position)
        except Exception:
            traceback.print_exc()
            return response

        if action is not None:
            response.action.MergeFrom(action)

        return response

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    protos_pb2_grpc.add_StrategyEvaluatorServicer_to_server(
        StrategyEvaluatorServer(), server)
    server.add_insecure_port('[::]:50051')
    server.start()
    server.wait_for_termination()

if __name__ == '__main__':
    serve()
