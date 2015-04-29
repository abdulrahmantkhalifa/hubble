initial poc for some networking tricks
======================================

![example](https://docs.google.com/drawings/d/1JyMhktMNzaOeIgbkAcNK7DdkV0M17eKoYzw1yxwCOlQ/pub?w=960&h=720)
click [here](https://docs.google.com/drawings/d/1JyMhktMNzaOeIgbkAcNK7DdkV0M17eKoYzw1yxwCOlQ/edit?usp=sharing) to edit picture

* 212.3.247.26:80 -> ends at rest app server on port 80
* 212.3.247.26:9000 -> ends at redis 1 on port 9000
* 212.4.1.1:80 -> ends at rest app server on port 80
* 212.4.1.1:9000 -> ends at redis 1 on port 9000
* 192.168.2.1:9000 -> ends at redis 1 on port 9000
* 192.168.2.1:80 -> ends at app server on port 80
* 192.168.1.10:9001 -> ends at redis 2 on port 9000
* 192.168.1.10:3306 -> ends at mysql1 on port 3306

features
--------
- SSL encryption possible

remarks
--------
* remote side is behind NAT & NO portforwarding on router at that location !!! so remote pc needs to keep connection open to pub GW to make sure connections can get back
* remote appserver can connect to mysql1 like it would connect to a local mysql
* the director is used to keep track of map of services and what needs to go through where to access something
* directory is rest web service using redis as backend store 
  * store the services map in redis on director
* the GW is only 1 binary 
  * command line options
    * the directory server to connect to (in 2nd phase will use raft for that)
    * name of GW
    * link to ssl key???
* the directory server reads a config file (toml) and reloads automatically when config file changes
  * config file identifies all mappings (see example above)
* each GW will do connection tests if it can reach the different GW and if not directly how it needs to go through other GW´s to get to the destination (means one or more of GW´s will serve as proxy to other GW (to be able to go through NAT GW/FW))
* each service e.g. redis 2 is serviced by only 1 GW (for now), other GW´s will go over this GW directly if they can connect it directly, if not they go over proxy GW to get to right GW. 
* each gateway need to monitor local ports of local services e.g. GW1 will warn director if redis on 192.168.1.10:9000 no longer responds
* make sure that whatever happens at remote side e.g. network goes down, nat gateway out for 1 h, the gw keeps on reconnecting and automatically is available again when network is back

need specs
----------
- how dow we best do the security & key exchange?
- spec config file

relevant work to look at
------------------------
- https://www.consul.io/
  - uses raft
  - uses dns to find relevant gateways (think this is good idea)
- https://github.com/zettio/weave
- there are quite some ssh forward examples in golang
