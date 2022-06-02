#!/bin/sh
vtysh -c "conf t" \
          -c "router bgp 300" \
          -c "bgp router-id 3.3.3.3" \
          -c "neighbor 10.1.0.2 remote-as 100" \
          -c "network 10.3.0.0/24"
