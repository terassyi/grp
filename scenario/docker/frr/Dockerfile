FROM frrouting/frr:v7.5.0

RUN apk update && apk add tcpdump iperf
RUN mkdir -p /tmp/data

COPY daemons /etc/frr/
COPY vtysh.conf /etc/frr/
COPY docker-start /tmp/docker-start
COPY udp_multicast_sender.py /

ENTRYPOINT [ "/tmp/docker-start" ]
