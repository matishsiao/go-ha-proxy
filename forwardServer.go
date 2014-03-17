package main
	
	  import (
	    "net"
	    "fmt"    
	    "io"
	    "time"
	    "container/list"
	    "strings"
	  )
	
	
	
	var (
		count int = 0			
	)
	type ForwardServer struct {
		ClientList *list.List	
		srvProxy Proxy
		localListener net.Listener
		Run bool	
	}
	
	type Client struct {
		Conn net.Conn
		DstIndex int
		RemoteAddr string
	}
	
	func (fs *ForwardServer) CheckHealth(connType string,uri string) (bool, int){  	
	    //fmt.Printf("checkHealth:Type:%s Addr:%s\n",connType,uri)
	    conn,err := net.Dial(connType,uri)
	    errCode := 0
	    
	    if err != nil {	        
	        errCode = 1
	        return false,errCode
	    }
	    
	    defer conn.Close()
	    
	    conn.SetReadDeadline(time.Now().Add(5 * time.Second))	    
	    
	    switch connType {
	    	case "udp":
	    		if conn != nil {
	    			conn.Write([]byte("checkHealth"))
		    		buffer := make([]byte, 1024)
					_, err := conn.Read(buffer)				
					
					if err != nil {
						errCode = 2
						if strings.Index(err.Error(), "timeout") != -1 {
							errCode = 0
							return true,errCode
						}
						return false,errCode
					}
					
					return true,errCode	
			}else {
				errCode = 3
				return false,errCode
			}
			break
	    }
		return true,errCode
	}
	
	
	func CheckTimeout(localConn net.Conn,proxy Proxy) {
		if proxy.KeepAlive != 0 {
			localConn.SetReadDeadline(time.Now().Add(time.Duration(proxy.KeepAlive) * time.Second))
		}
	}
	
	
	func (fs *ForwardServer) GetClientElement(RemoteAddrs string) *list.Element {
		//RemoteAddrs := _RemoteAddr
		RemoteAddr := RemoteAddrs[:strings.Index(RemoteAddrs,":")]
		RemotePort := RemoteAddrs[strings.Index(RemoteAddrs,":"):]
		fmt.Printf("RAddr:%s RPort:%s\n",RemoteAddr,RemotePort)
		for e := fs.ClientList.Front(); e != nil; e = e.Next() {				
			if e.Value.(*Client).Conn.RemoteAddr().String() == RemoteAddr {					
				return e
		    }				
		}
		return nil
	}
	
	
	func (fs *ForwardServer) GetClient(RemoteAddrs string) *Client {
		//RemoteAddrs := _RemoteAddr
		RemoteAddr,RemotePort := GetRemoteAddrInfo(RemoteAddrs)		
		fmt.Printf("RAddr:%s RPort:%s\n",RemoteAddr,RemotePort)
		for e := fs.ClientList.Front(); e != nil; e = e.Next() {
			clientRAddr,_ := GetRemoteAddrInfo(e.Value.(*Client).Conn.RemoteAddr().String())		
			if clientRAddr == RemoteAddr {			
				return e.Value.(*Client)
		    }				
		}
		return nil
	}
	
	func GetRemoteAddrInfo(RemoteAddrs string) (string,string) {
		RemoteAddr := RemoteAddrs[:strings.Index(RemoteAddrs,":")]
		RemotePort := RemoteAddrs[strings.Index(RemoteAddrs,":"):]
		return RemoteAddr,RemotePort
	}
	
	
	func (fs *ForwardServer) Forward(localConn net.Conn,serverAddrString string) {	
	    // Setup server Conn	    
	    count++
		srvConn, err := net.Dial(fs.srvProxy.Mode, serverAddrString)
		if err != nil {
			fmt.Printf("forward Err: %v\n", err)
			return
		}
		//fmt.Printf("connType:%s serverAddr:%s fowarding[%d] \n",connType,serverAddrString,count)
		// Copy localConn.Reader to sshConn.Writer		
		CheckTimeout(localConn,fs.srvProxy)
		if srvConn != nil {		    	
		    go func() {
			   _, err := io.Copy(srvConn, localConn)
			   if err != nil {
			   		if *configInfo.Debug {
			        	fmt.Printf("io.Copy S2L failed: %v\n", err)
			        }
					srvConn.Close()
			        localConn.Close()
			        return
			   }	     
			}()
			
			// Copy srvConn.Reader to localConn.Writer
			go func() {
			   _, err := io.Copy(localConn, srvConn)				   
			   if err != nil {
			   		if *configInfo.Debug {
			       		fmt.Printf("io.Copy L2S failed: %v\n", err)
			       	}
			        srvConn.Close()
			        localConn.Close()
			        return
			   }
			}()
			//defer srvConn.Close()
		}
		//defer localConn.Close()
	}
	
	
	func (fs *ForwardServer) Check()  {
		for fs.Run {
			//fmt.Printf("FS Check:%s\n",fs.srvProxy.Name)
			for k, dstObj := range fs.srvProxy.DstList {	
					
				if dstObj.Check {		
					fs.srvProxy.DstList[k].Health,_ = fs.CheckHealth(fs.srvProxy.Mode,fs.srvProxy.GetDstAddr(k))					
				} else {
					fs.srvProxy.DstList[k].Health = true
				}
				//fmt.Printf("Check Mode:%s DstAddr:%s Health:%v SrvHealth:%v\n",fs.srvProxy.Mode,fs.srvProxy.GetDstAddr(k),dstObj.Health,fs.srvProxy.DstList[k].Health)
			} 
			time.Sleep(time.Duration(fs.srvProxy.CheckTime) * time.Second)
			
		}
	}
	
	
	func (fs *ForwardServer) GetHealthNode(DstIndex int) int {
		healthIndex := -1
		if !fs.srvProxy.DstList[DstIndex].Health {			
		    //從目前之後的節點找出健康的節點使用
		    for i := fs.srvProxy.Index; i < fs.srvProxy.DstLen;i++ {
		    	if fs.srvProxy.DstList[i].Health {
		        	healthIndex = i
		        	break
		        }
		   	} 
		        				
		   	//如果目前之後的節點沒有健康的,則從全部節點重新找一次
		   	if healthIndex == -1 {
		   		for i := 0; i < fs.srvProxy.DstLen;i++ {
					if fs.srvProxy.DstList[i].Health {
						healthIndex = i
						break
					}
		    	}
		    }
		       				
		    //如果都沒有健康的節點,則使用第一個節點
		    if healthIndex == -1 {
		    	healthIndex = 0
		    }
		        				
			fs.srvProxy.Index = healthIndex		        				
		} else {
			healthIndex = DstIndex
		}
		return healthIndex
	}
	
	
	func (fs *ForwardServer) TurnToNode(localConn net.Conn) {
			
		switch fs.srvProxy.Type {
			case "LeastConn":
		
			case "Weight":	        				        		
			       		

			case "Source":
				client := fs.GetClient(localConn.RemoteAddr().String())
				if client == nil {
					fs.srvProxy.Index = fs.srvProxy.Counter % fs.srvProxy.DstLen
					fs.srvProxy.Counter++
					client = new(Client)		        			     				
					client.DstIndex = fs.GetHealthNode(fs.srvProxy.Index)
					client.Conn = localConn							
					fs.ClientList.PushBack(client)							
				} else {
					client.DstIndex = fs.GetHealthNode(client.DstIndex)
				}
				fs.srvProxy.DstList[client.DstIndex].Counter++
				if *configInfo.Debug {
					fmt.Printf("DstAddr:%s Remote:%s DstIndex:%d Client:%v\n",fs.srvProxy.GetDstAddr(client.DstIndex),localConn.RemoteAddr().String(),client.DstIndex,client)
				}
				go fs.Forward(localConn,fs.srvProxy.GetDstAddr(client.DstIndex))
			case "RoundRobin":
				fs.srvProxy.Index = fs.srvProxy.Counter % fs.srvProxy.DstLen
				fs.srvProxy.Counter++
				DstIndex := fs.GetHealthNode(fs.srvProxy.Index)
				fs.srvProxy.DstList[DstIndex].Counter++
				if *configInfo.Debug {
					fmt.Printf("DstAddr:%s Remote:%s\n",fs.srvProxy.GetDstAddr(DstIndex),localConn.RemoteAddr().String())
				}		        		
				go fs.Forward(localConn,fs.srvProxy.GetDstAddr(DstIndex))			
		}
	}
	
	
	func (fs *ForwardServer) Reload(srvProxy Proxy) {
		fs.srvProxy = srvProxy
		fmt.Printf("FS:%s reloaded.\n",fs.srvProxy.Name)
		
	}
	
	
	func (fs *ForwardServer) Stop() {
		fs.Run = false
		if fs.localListener != nil {
			fs.localListener.Close()
		}
		//fmt.Printf("FS:%s ready to stop.%v\n",fs.srvProxy.Name,fs.Run)
	}
	
	func (fs *ForwardServer) Listen(srvProxy Proxy)  {	
		if *configInfo.Debug {	
			fmt.Printf("forwardServer connType:%s serverAddr:%s Type:%s\n",srvProxy.Mode,srvProxy.GetSrcAddr(),srvProxy.Type)
		}
		//已經使用var 宣告則物件已建立,不需要再用new
		//FS := new(ForwardServer)
		fs.ClientList = list.New()
		fs.Run = true
		fs.srvProxy = srvProxy
		fs.srvProxy.Counter = 0
		if fs.srvProxy.CheckTime == 0 {
			fs.srvProxy.CheckTime = 5
		}
		if fs.srvProxy.Mode == "tcp" || fs.srvProxy.Mode == "http" || fs.srvProxy.Mode == "health" {
			fs.srvProxy.DstLen = len(fs.srvProxy.DstList)
			switch fs.srvProxy.Mode {
				case "health":
					go fs.Check()					    
				default:
					localListener, err := net.Listen("tcp", fs.srvProxy.GetSrcAddr())
					fs.localListener = localListener
				    if err != nil {
				    	if *configInfo.Debug {
				        	fmt.Printf("net.Listen failed: %v\n", err)
				        }
				        return
				    }
				    
				    //確認是否要檢查遠端主機
				    for _, dstObj := range fs.srvProxy.DstList {	
				    	if dstObj.Check {
				    		go fs.Check()	    		
				    		break
				    	}
					} 
				     		
				    //監聽Port 		    
			    	for {
					    // Setup localConn (type net.Conn)
					    localConn, err := fs.localListener.Accept()
					    if err != nil {
					    	if *configInfo.Debug {
					       		fmt.Printf("listen.Accept failed: %v\n", err)
					       	}
					    	break
					    }
					    
					    fs.TurnToNode(localConn)  
				    }
				    
				    fmt.Printf("FS:%s is stoped.\n",fs.srvProxy.Name)
			} 
		    
		} else {
			fmt.Printf("Unsupport mode:%s,listen failed.\n",fs.srvProxy.Mode)
		}
	}

