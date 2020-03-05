FROM ubuntu:18.06
ADD ./kwont /kwont
ENTRYPOINT ["/kwont"]
