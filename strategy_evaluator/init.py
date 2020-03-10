from concurrent import futures
import time
import math
import logging
import grpc

import protos_pb2
import protos_pb2_grpc

class StrategyEvaluatorServer(protos_pb2_grpc.StrategyEvaluatorServicer):
    def Evaluate(self, request: protos_pb2.EvaulateStrategyRequest, context):
        response = protos_pb2.EvaluateStrategyResponse()
        response.action.open_position.price = 123

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