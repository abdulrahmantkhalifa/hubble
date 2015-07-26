[![build status](http://gitci.aydo.com/projects/1/status.png?ref=master)](http://gitci.aydo.com/projects/1?ref=master)

Hubble
======

Hubble allows clients behind Firewalled natted networks to reach services behind different
firewalled natted services by proxying the traffic over websockets. Websockets are usually
allowed to reach outside the natted network since they are http protocol.

# How to use
To have a working setup you need to run the following:
1- A proxy server, this one must be reachable from both natted networks (on the public internet)
2- 2 agents, one for each natted network.

Steps:
```sh
go get git.aydo.com/0-complexity/hubble
```

## Demo environemt
For testing you still can run the proxy and the 2 agents on the same machine as following

### Running proxy
```sh
cd $GOPATH/src/git.aydo.com/0-complexity/hubble/main
go run proxy.go
```

By default proxy will start listing on port 8080. You can change that with the `-listen` option (ex: go run proxy.go -listen=127.0.0.1:80)

### Running agent1
```sh
cd $GOPATH/src/git.aydo.com/0-complexity/hubble/main
go run agent.go -url=ws://localhost:8080 -name=agent1 2222:agent2:127.0.0.1:22
```

The forwarding rule reads as "Listen on local port `2222` and forward connection to that port to over agent2 to machine 127.0.0.1:22 (in agent2 network)"
You can add as many forwarding rules as you want

If you want the application to dynamically choose an unused local port to listen on, you may specify `0` as the local port. The application will log the actual chosen port to the console.

### Running agent2
```sh
cd $GOPATH/src/git.aydo.com/0-complexity/hubble/main
go run agent.go -url=ws://localhost:8080 -name=agent2
```

Note that agent2 doesn't define forwarding rules, which means it only accepts incoming traffic from the proxy. agent1 also accepts incoming traffic, but also allows outcoming traffic (on port 2222)

### Testing the setup
Simply do:
```sh
ssh -p 2222 localhost
```