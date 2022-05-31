#!/bin/sh

vtysh -c "conf t" \
          -c "router bgp 200" \
          -c "bgp router-id 2.2.2.2" \
          -c "neighbor 10.0.0.2 remote-as 100"
