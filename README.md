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
go get github.com/Jumpscale/hubble
```

## Demo environemt
For testing you still can run the proxy and the 2 agents on the same machine as following

### Running proxy
```sh
cd $GOPATH/src/github.com/Jumpscale/hubble/main
go run proxy.go
```

By default proxy will start listing on port 8080. You can change that with the `-listen` option (ex: go run proxy.go -listen=127.0.0.1:80)

### Running agent1
```sh
cd $GOPATH/src/github.com/Jumpscale/hubble/main
go run agent.go -url=ws://localhost:8080 -name=agent1 2222:agent2:127.0.0.1:22
```

The forwarding rule reads as "Listen on local port `2222` and forward connection to that port to over agent2 to machine 127.0.0.1:22 (in agent2 network)"
You can add as many forwarding rules as you want

If you want the application to dynamically choose an unused local port to listen on, you may specify `0` as the local port. The application will log the actual chosen port to the console.

If you have enabled authorization you might have to specify a key to open a tunnel. This key has to be specified on the command line.
```sh
go run agent.go -url=ws://localhost:8080 -name=agnet1 2222:myconnectiontoken@agent2:127.0.0.1:22
```

### Running agent2
```sh
cd $GOPATH/src/github.com/Jumpscale/hubble/main
go run agent.go -url=ws://localhost:8080 -name=agent2
```

Note that agent2 doesn't define forwarding rules, which means it only accepts incoming traffic from the proxy. agent1 also accepts incoming traffic, but also allows outcoming traffic (on port 2222)

### Testing the setup
Simply do:
```sh
ssh -p 2222 localhost
```

# Authorizing tunnels
You might want to grant or deny requests from clients to open a tunnel. This is managed by the authorization module. There are currently three types:

 - **Grant all**: Grant all requests to open a tunnel (use the `-authgrant` flag)
 - **Deny all**: Deny all requests to open a tunnel (use the `-authdeny` flag)
 - **Lua module**: Use an external Lua script to handle authorization requests (use the `-authlua script.lua` flag)

## External Lua module
The external Lua modules are ran using [gopher-lua](https://github.com/yin/gopher-lua), with access to the following modules:

 - [**gluahttp**](https://github.com/cjoudrey/gluahttp): HTTP request module for gopher-lua
 - [**gopher-json**](https://github.com/layeh/gopher-json): Simple JSON encoder/decoder for gopher-lua

Your script must hava a global function `connect` that accepts a `request` object and returns either `true` or `false`.

```lua
function connect(request)
    -- Decline connection requests.
    return false
end
```

The `request` object has the following methods:

 - `request:ip()`: destination IP address in the destination network *(string)*
 - `request:port()`: destination port *(number)*
 - `request:gatename()`: gatename of the destination network *(string)*
 - `request:key()`: the key specified in the connection handshake *(string)*

A simple example using an HTTP request to authorize a connection can be found in [example.lua](auth/example.lua).
