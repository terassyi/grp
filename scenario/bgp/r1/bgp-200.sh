#!/bin/sh
vtysh -c "conf t" \
          -c "router bgp 200" \
          -c "bgp router-id 2.2.2.2" \
          -c "neighbor 10.0.0.2 remote-as 100" \
          -c "neighbor 10.2.0.3 remote-as 400" \
          -c "network 10.2.0.0/24" \
          -c "network 10.2.4.0/24"
