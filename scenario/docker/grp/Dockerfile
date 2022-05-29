FROM golang:1.18

RUN apt update -y && \
	apt install -y vim tcpdump iproute2 iputils-ping systemd iperf git && \
	git clone https://github.com/terassyi/grp.git && \
	cd grp && \
	go build -o grpd && \
	mv grpd /usr/local/bin

WORKDIR /go/grp

CMD [ "bash" ]
