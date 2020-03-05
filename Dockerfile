FROM alpine:latest
RUN apk --no-cache --update add ca-certificates
ADD ./kwont /kwont
ENTRYPOINT ["/kwont"]
