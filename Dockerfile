FROM alpine:3.16.3

ADD ./heartbeatctl /usr/local/bin/heartbeatctl

ENTRYPOINT ["/usr/local/bin/heartbeatctl"]
