Introduction:

  This GoHAProxy is simple HAProxy by Golang.

  It's support tcp proxy and you can use roundrobin or source mode to HA.
  

Install:

  go get github.com/matishsiao/GoHAProxy

  go test

  go build

Configuration:
  all configs in haproxy.json
  ProxyList:all proxy data.
  Proxy:
      Src:source ip or domain.
      SrcPort:source port.
      Mode:tcp (http and health will add in next version.)
      Type:RoundRobin,Source(Weight will add in next version.)
      KeepAlive:1 second (keep alive server connection.)
      CheckTime:1 second (default 5 seconds.)
      DstList:Destination server list
	DstNode:
          Name:server name
          Dst:server ip or domain
          DstPort:destination port
          Weight:not use(when Weight HA mode done,will use this arg)
          Check:true or false(if you set false,the GoHAProxy will set this server allways health.)
