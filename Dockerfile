FROM alpine:3.22.0

ADD ./heartbeatctl /usr/local/bin/heartbeatctl

ENTRYPOINT ["/usr/local/bin/heartbeatctl"]
