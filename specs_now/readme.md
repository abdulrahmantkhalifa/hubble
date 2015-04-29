initial poc for some networking tricks
======================================

![example](https://docs.google.com/drawings/d/1CkCJhaFLX4KoU2_Ay1MpRGey4MzjTqcb8oqxPp1LYvg/pub?w=960&h=720)
click [here](hhttps://docs.google.com/drawings/d/1CkCJhaFLX4KoU2_Ay1MpRGey4MzjTqcb8oqxPp1LYvg/edit) to edit picture

this is an example where we ask the gw to connect 2 tcp ports
so the python api client can talk to the remote server and both are behind NAT.

## goal

- create a prototype gateway can be in lua, python or golang.
- which allows us to connect ports to be connected even over NAT firewalls

## principle

- network
	- 192.168.10.0/24 is behind GW1
	- 10.10.10.0/24 is behind GW2
- goal connect port from GW2 to GW1 lan ip/port
- python api asks GW2 connect port on GW1 to a remote port behind GW1
	-  connect(remotegw="gw1",addr=192.168.10.12,port=22,localport=2022)
	-  localport is port at internal side of GW2 which will be forwarded
- use http(s) to GW3-4-5  
	- use websockets
	- multiplex sockets over the websockets
	- DO NOT DO TCP over Websockets or TCP
- GW 3-4-5
	- are the central GW's which serve as proxy (terminating http)
	- so there is no other traffic than http(s) over internet and support for NAT is in
- put some filtering in on GW3-4-5, who can talk with who?
	
## info & examples

- http://blog.magiksys.net/software/tcp-proxy-reflector#tcpr_install
- https://github.com/benoitc/tproxy
- https://github.com/SnabbCo/snabbswitch

## requirements

- work on openwrt/linux/windows/mac (proto linux)
- the central GW's can be on linux only?

## remarks

- its basically some smart forwarding of socket messages over websockets




