package main

	import (
	    //"net"
	    "fmt"
	    //"strings"
	    "time"
	    "encoding/json"
    	"io/ioutil"
    	"os"
	)
	
	/*func test() {
		for i := 0;i < 1000;i++ {
		  status,err := checkHealth("udp","127.0.0.1:9000")
	          fmt.Printf("uri:%s mode:%s status:%v err:%d\n","127.0.0.1:9000","udp",status,err)
	          time.Sleep(100 * time.Millisecond)
		}
	}*/

	func main() {
	   /* status,err := checkHealth("udp","8.8.8.8:53")
	    fmt.Printf("uri:%s mode:%s status:%v err:%d\n","8.8.8.8:53","udp",status,err)*/
	    file, e := ioutil.ReadFile("./haproxy.json")
	    if e != nil {
	        fmt.Printf("File error: %v\n", e)
	        os.Exit(1)
	    }
	    fmt.Printf("%s\n", string(file))
	    var haConfig HAConfig
	    json.Unmarshal(file, &haConfig)
	    fmt.Printf("Results: %v\n", haConfig)
	    fmt.Printf("main forwardServer\n")
	    for k,proxy := range haConfig.Configs.ProxyList {
	    	fmt.Printf("Proxy[%v]: %v\n", k,proxy)
	    	FS := new (ForwardServer)
	    	go FS.Listen(proxy)	
	    }
	    /*udpFS := new (ForwardServer)
	    tcpFS := new (ForwardServer)
	    
	    go tcpFS.Listen("tcp",":9000","10.7.9.53:80")
	    */
	    for {
			time.Sleep(100 * time.Millisecond)
	    }
	}
