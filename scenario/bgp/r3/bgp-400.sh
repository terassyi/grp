#!/bin/sh
vtysh -c "conf t" \
          -c "router bgp 400" \
          -c "bgp router-id 4.4.4.4" \
          -c "neighbor 10.2.0.2 remote-as 200" \
          -c "network 10.4.0.0/24"
