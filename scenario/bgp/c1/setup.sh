#!/bin/sh

ip route del default
ip route add default via 10.4.0.2
