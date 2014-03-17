package main

	import (	
	    "fmt"	    
	    "time"
	    "encoding/json"
    	"io/ioutil"
    	"os"
    	"flag"
	)
	var version string = "0.0.2"
	var proxyServer ProxyServer
	var configInfo ConfigInfo
	var help *bool
	func main() {
		//fmt.Printf("GoHAProxy version:%s\n",version)
		configInfo.FileName = flag.String("config","./config.json","set config file path.")
		configInfo.Debug = flag.Bool("debug",false,"show debug trace message.")		
		configInfo.Version = flag.Bool("version",false,"GoHAProxy version.")
		help = flag.Bool("help",false,"Show help information.")
		flag.Parse()
		
		if *help {
			fmt.Printf("GoHAProxy is simple HAProxy by Golang,You can use this to load balance TCP or check health status.\n")
			fmt.Println("GoHAProxy Version", version)
			fmt.Printf("Monitor server:local ip:8080, Default value:[:8080]\n")
			fmt.Printf("-config    Set cofing file path. Default value:%v\n", *configInfo.FileName)
			fmt.Printf("-debug     Show debug trace message. Default value:%v\n", *configInfo.Debug)
			fmt.Printf("-version   Show GoHAProxy version.\n")
			fmt.Printf("-help      Show help information.\n")
			os.Exit(0)
		}
		
		
		if *configInfo.Version {
			fmt.Println("GoHAProxy Version", version)
			os.Exit(0)
		}
		
		fmt.Printf("GoHAProxy FileName:%s\n",*configInfo.FileName)
		
	    haConfig := loadConfigs()	    	    
	    
	    for _,proxy := range haConfig.Configs.ProxyList {	    
	    	FS := new (ForwardServer)
	    	proxyServer.ServerList = append(proxyServer.ServerList, FS)
	    	//proxyServer.ServerList[k] = FS
	    	go FS.Listen(proxy)	
	    }
	    
	    go Monitor()
	    for {
	    	/*for k,proxy := range proxyServer.ServerList {
	    		fmt.Printf("Proxy[%v]: %v\n", k,proxy.srvProxy)
	    	}*/			
			configWatcher()			
			time.Sleep(500 * time.Millisecond)
	    }
	}
	
	
	func configWatcher() {
		file, err := os.Open(*configInfo.FileName) // For read access.
		if err != nil {
			fmt.Println(err)
		}
		info,err := file.Stat()
		if err != nil {
			fmt.Println(err)
		}
		if configInfo.Size == 0 {
			configInfo.Size = info.Size()
			configInfo.ModTime = info.ModTime()
		} 
		
		
		
		if info.Size() != configInfo.Size || info.ModTime() != configInfo.ModTime {
			fmt.Printf("Config changed.Reolad.\n")
			configInfo.Size = info.Size()
			configInfo.ModTime = info.ModTime()	
			haConfig := loadConfigs()
			lastkey := 0
			//檢查有沒有移除掉的設定有的話就移除
		    for k,sProxy := range proxyServer.ServerList {
		    	oldProxy := true
				for _,proxy := range haConfig.Configs.ProxyList {
					if *configInfo.Debug { 
						fmt.Printf("Delete PName:%s srvProxyName:%s \n",proxy.Name,sProxy.srvProxy.Name)
					}
					if proxy.Name == sProxy.srvProxy.Name {
						oldProxy = false
						break
					}
			    }
			    if k > lastkey {
			    	lastkey = k
			    }
			    if oldProxy {
			    	if *configInfo.Debug {
			    		fmt.Printf("Delete Proxy[%v]: %v\n", k,sProxy.srvProxy.Name)
			    	}
			    	sProxy.Stop()
			    	//proxyServer.ServerList = append(proxyServer.ServerList[:k], proxyServer.ServerList[k+1:])
			    	proxyServer.ServerList = proxyServer.ServerList[:k+copy(proxyServer.ServerList[k:], proxyServer.ServerList[k+1:])]
			    	//proxyServer.ServerList = copy(proxyServer.ServerList[k:], proxyServer.ServerList[k+1:])
			    	
			    	//delete(proxyServer.ServerList, k)
			    	//time.Sleep(1 * time.Second)
			    }
		    }
		    
			//檢查有沒有新的設定
			for _,proxy := range haConfig.Configs.ProxyList {
				newProxy := true
				for k,sProxy := range proxyServer.ServerList {
					if *configInfo.Debug {
						fmt.Printf("Check Add PName[%d]:%s srvProxyName:%s \n",k,proxy.Name,sProxy.srvProxy.Name)
					}
					if proxy.Name == sProxy.srvProxy.Name {
						proxyServer.ServerList[k].Reload(proxy)
						newProxy = false
						break					
					}
		    	}	
				if newProxy {
					FS := new (ForwardServer)
		    		proxyServer.ServerList = append(proxyServer.ServerList, FS)
		    		//proxyServer.ServerList[lastkey+1] = FS
		    		if *configInfo.Debug {
		    			fmt.Printf("Add New Proxy: %v\n", proxy)
		    		}
		    		go FS.Listen(proxy)
				}
		    }
		    
		    
					
		}
		defer file.Close()
	}
	
	
	func loadConfigs() HAConfig {
		file, e := ioutil.ReadFile(*configInfo.FileName)
	    if e != nil {
	        fmt.Printf("Load GoHAProxy config error: %v\n", e)
	        os.Exit(1)
	    }	    
	    var haConfig HAConfig
	    json.Unmarshal(file, &haConfig)	    
	    return haConfig
	}

	