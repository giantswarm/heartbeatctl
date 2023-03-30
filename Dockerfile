FROM alpine:3.17.3

ADD ./heartbeatctl /usr/local/bin/heartbeatctl

ENTRYPOINT ["/usr/local/bin/heartbeatctl"]
