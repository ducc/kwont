FROM ubuntu:18.04
ADD ./kwont /kwont
ENTRYPOINT ["/kwont"]
