FROM alpine:latest
ADD ./kwont /kwont
ENTRYPOINT ["/kwont"]
