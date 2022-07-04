FROM golang:1.18

RUN apt update -y && \
	apt install -y vim tcpdump iproute2 iputils-ping systemd iperf git make protobuf-compiler

RUN go install github.com/cweill/gotests/gotests@latest && \
	go install github.com/fatih/gomodifytags@latest && \
	go install github.com/josharian/impl@latest && \
	go install github.com/haya14busa/goplay/cmd/goplay@latest && \
	go install honnef.co/go/tools/cmd/staticcheck@latest && \
	go install golang.org/x/tools/gopls@latest

# RUN git clone https://github.com/terassyi/grp.git && \
# 	cd grp && \
# 	go build -o grpd && \
# 	mv grpd /usr/local/bin

WORKDIR /go/grp

# RUN go mod download

CMD [ "bash" ]
