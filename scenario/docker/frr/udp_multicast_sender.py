#! /usr/bin/python3

import socket
import sys
import time

args = sys.argv
src = ""
dst = ""
port = 0
count = 0
src_next = False
dst_next = False
port_next = False
count_next = False

for arg in args:
	if arg == "--src" or arg == "-s":
		src_next = True
		continue
	elif arg == "--dst" or arg == "-d":
		dst_next = True
		continue
	elif arg == "--port" or arg == "-p":
		port_next = True
		continue
	elif arg == "--count" or arg == "-c":
		count_next = True
		continue

	if src_next:
		src = arg
		src_next = False
	elif dst_next:
		dst = arg
		dst_next = False
	elif port_next:
		port = int(arg)
		port_next = False
	elif count_next:
		count = int(arg)
		count_next = False
		

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM, socket.IPPROTO_UDP)
sock.setsockopt(socket.IPPROTO_IP, socket.IP_MULTICAST_LOOP, 0)
sock.setsockopt(socket.IPPROTO_IP, socket.IP_MULTICAST_IF, socket.inet_aton(src))

for i in range(count):
	message = "UDP Multicast test {0}".format(i).encode('utf-8')
	sock.sendto(message, (dst, port))
	time.sleep(1)
