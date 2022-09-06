FROM alpine:3.16.2

ENV PATH /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

ADD ./heartbeatctl /usr/local/bin/heartbeatctl
ENTRYPOINT ["/usr/local/bin/heartbeatctl"]