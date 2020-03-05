FROM ubuntu:18.04
RUN apk --no-cache --update add ca-certificates
ADD ./kwont /kwont
ENTRYPOINT ["/kwont"]
