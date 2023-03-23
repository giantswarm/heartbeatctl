FROM alpine:3.17.2

ADD ./heartbeatctl /usr/local/bin/heartbeatctl

ENTRYPOINT ["/usr/local/bin/heartbeatctl"]
