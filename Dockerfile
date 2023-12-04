FROM alpine:3.18.5

ADD ./heartbeatctl /usr/local/bin/heartbeatctl

ENTRYPOINT ["/usr/local/bin/heartbeatctl"]
