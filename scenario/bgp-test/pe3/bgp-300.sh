#!/bin/sh

vtysh -c "conf t" \
          -c "router bgp 300" \
          -c "bgp router-id 3.3.3.3" \
          -c "neighbor 10.0.2.4 remote-as 200"
