FROM python:3.6.4
COPY . /app
WORKDIR /app

RUN pip install grpcio-tools
RUN pip install pandas

EXPOSE 50051
CMD python ./server.py