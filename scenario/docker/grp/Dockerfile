FROM golang:1.18

RUN apt update -y && \
	apt install -y vim tcpdump iproute2 iputils-ping systemd iperf

COPY init.sh /tmp/init.sh

CMD [ "bash" ]
