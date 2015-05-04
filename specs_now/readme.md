initial poc for some networking tricks
======================================

![example](https://docs.google.com/drawings/d/1CkCJhaFLX4KoU2_Ay1MpRGey4MzjTqcb8oqxPp1LYvg/pub?w=960&h=720)
click [here](https://docs.google.com/drawings/d/1CkCJhaFLX4KoU2_Ay1MpRGey4MzjTqcb8oqxPp1LYvg/edit) to edit picture

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


initial implementation details for the MVP
==========================================

![Sequence Diagram](http://www.plantuml.com/plantuml/img/RLB1JW8n4BtlLqouYv4Z1nEYKJaOITHuuJ9qXwMXp3JjhClwzKvPL20sQJlDp7jl-bfqAWdUCoN03AtjLSIatlc8h32QDMJRpQXaiSGtP_b7LEgmBzcc-myv-KDEgfMqN6FgOVHAwTCxEYJp45TLawIDC6Ul7eF_yjp0KRuQfE7grcIcy8HSvmrk2Im0R7L3x1sg5xweV8d4yF4AJfZ97Ge0WaY41xiseTkKrDetRQ8Qrf8wJBMLsFWZ6g8ZMH270Q8aNN9Hxz1h0Pv8P2DWKO90QpsGtVUCE-z-138uP5WHP3N7p7hvj2LHLfWjveOW8ouCGrYLEHwFg_8ybmsPkLQQTXYM_7Qtk9ulJYzxxiPp6zJ7G7k8s0V1Uico1aFPsMsjriwbauvMaCLjIQkc-nMSJE7UvuXgfjMe22fWpY-vfkmRjfzGGpFyXq5tKg37rvt28ic-Fn4kGaF3tm00)

## Components
- **Agent** which runs inside the natted network, it acts as the local proxy for the remotely natted services. To have a working setup, you need at least 2 **agents** that runs inside the 2 networks you want to connect.
- **Gateway** which runs in the cloud, it works as the main central dispatching point. It's responsable for forwarding traffic from one agent to the other. It also should manage authentication and authroization so only authenticated agents can connect to the Gateway, and only allowed agents can forward traffic to the designated agents.

### Agent
Agent, when started it will connect to the configured proxy server over websockets. The handshake process will include the following steps:
- Authentication, only authenticated gateways will proceed to the active state, otherwise a decent error message will be reported back to the agent and connection will be terminated.
- Registration, The **Agent** will identify it's usable name (agentname), The agent name must be unique. The agentname will be used by other agents when forwarding connections.
- When the registration process is complete, the agent can now proceed with opening the local ports which maps to remote services. A typical port forwarding configuration has the following structure:
```javascript
 {
 	local: 'local port number',
 	gateway: 'remote gatename',
 	ip: 'remote server ip (the private IP of the service in the remote network)'
 	remote: 'remote port number'
 }
 ```
-- The agent should provide a RestAPI to dynamically open and close forwardings.
- Internally, the agent must keep track of the connected sockets so it can route the received traffic back the correct socket.

Agents also are used as the entry point for internal services, so when agent receives a new connection to an internal server, the server is first checked agains a white list to see if connection to this service is allowed, and if yes, the connection is establised and traffic is routed to the other end as desciped above.

## Protocol
TODO