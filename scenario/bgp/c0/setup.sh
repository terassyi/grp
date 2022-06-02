#!/bin/sh

ip route del default
ip route add default via 10.3.0.2
