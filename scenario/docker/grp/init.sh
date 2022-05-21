#! /bin/sh

set -e

# enable ip forwarding
sysctl -w net.ipv4.ip_forward=1

# del existing route table
