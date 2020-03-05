FROM scratch
ADD ./kwont /kwont
ENTRYPOINT ["/kwont"]
