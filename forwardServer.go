package main

import (
	"container/list"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
	"log"
)


type ForwardServer struct {
	ClientList    *list.List
	srvProxy      Proxy
	localListener net.Listener
	Run           bool
}

type Client struct {
	Conn       net.Conn
	DstIndex   int
	RemoteAddr string
}

func (fs *ForwardServer) CheckHealth(connType string, uri string) (bool, int) {
	if uri == "" {
		log.Println("Check health failed:uri is empty.")
		return false,0
	}
	conn, err := net.Dial(connType, uri)
	errCode := 0

	if err != nil {
		errCode = 1
		return false, errCode
	}
	
	

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
					return true, errCode
				}
				return false, errCode
			}

			return true, errCode
		} else {
			errCode = 3
			return false, errCode
		}
		break
	}
	conn.Close()
	return true, errCode
}

func (fs *ForwardServer) CheckTimeout(localConn net.Conn) {
	if fs.srvProxy.KeepAlive != 0 {
		localConn.SetReadDeadline(time.Now().Add(time.Duration(fs.srvProxy.KeepAlive) * time.Second))
	}
}

func (fs *ForwardServer) GetClientElement(RemoteAddrs string) *list.Element {
	RemoteAddr, _ := GetRemoteAddrInfo(RemoteAddrs)

	for e := fs.ClientList.Front(); e != nil; e = e.Next() {
		if e.Value.(*Client).Conn.RemoteAddr().String() == RemoteAddr {
			return e
		}
	}
	return nil
}

func (fs *ForwardServer) GetClient(RemoteAddrs string) *Client {	
	RemoteAddr, _ := GetRemoteAddrInfo(RemoteAddrs)

	for e := fs.ClientList.Front(); e != nil; e = e.Next() {
		clientRAddr, _ := GetRemoteAddrInfo(e.Value.(*Client).Conn.RemoteAddr().String())
		if clientRAddr == RemoteAddr {
			return e.Value.(*Client)
		}
	}
	return nil
}

func GetRemoteAddrInfo(RemoteAddrs string) (string, string) {
	RemoteAddr := RemoteAddrs[:strings.Index(RemoteAddrs, ":")]
	RemotePort := RemoteAddrs[strings.Index(RemoteAddrs, ":"):]
	return RemoteAddr, RemotePort
}

func (fs *ForwardServer) Forward(localConn net.Conn, index int) {
	// Setup server Conn	
	srvConn, err := net.Dial(fs.srvProxy.Mode, fs.srvProxy.GetDstAddr(index))
	if err != nil {
		fmt.Printf("forward Err: %v\n", err)
		return
	}
	fs.srvProxy.DstList[index].Connections++
	
	// Copy localConn.Reader to sshConn.Writer
	fs.CheckTimeout(localConn)
	fs.CheckTimeout(srvConn)
	
	if srvConn != nil {
		go func() {
			_, err := io.Copy(srvConn, localConn)
			if err != nil {
				if *configInfo.Debug {
					fmt.Printf("io.Copy S2L failed: %v\n", err)
				}
				fs.srvProxy.DstList[index].Connections--
				if fs.srvProxy.DstList[index].Connections < 0 {
					fs.srvProxy.DstList[index].Connections = 0
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
				fs.srvProxy.DstList[index].Connections--
				if fs.srvProxy.DstList[index].Connections < 0 {
					fs.srvProxy.DstList[index].Connections = 0
				}
				srvConn.Close()
				localConn.Close()
				return
			}
		}()	
	}	
}

func (fs *ForwardServer) Check() {
	for fs.Run {
		for k, dstObj := range fs.srvProxy.DstList {

			if dstObj.Check {
				fs.srvProxy.DstList[k].Health, _ = fs.CheckHealth("tcp", fs.srvProxy.GetDstAddr(k))
			} else {
				fs.srvProxy.DstList[k].Health = true
			}			
		}
		time.Sleep(time.Duration(fs.srvProxy.CheckTime) * time.Second)

	}
}

func (fs *ForwardServer) GetHealthNode(DstIndex int) int {
	healthIndex := -1
	if !fs.srvProxy.DstList[DstIndex].Health {
		//從目前之後的節點找出健康的節點使用
		for i := fs.srvProxy.Index; i < fs.srvProxy.DstLen; i++ {
			if fs.srvProxy.DstList[i].Health {
				healthIndex = i
				break
			}
		}

		//如果目前之後的節點沒有健康的,則從全部節點重新找一次
		if healthIndex == -1 {
			for i := 0; i < fs.srvProxy.DstLen; i++ {
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
		DstIndex := 0
		
		for k,Dst := range fs.srvProxy.DstList {
	    	if fs.srvProxy.DstList[DstIndex].Connections > Dst.Connections {
	    		DstIndex = k
	    	}	
		}
	    
	    DstIndex = fs.GetHealthNode(DstIndex)
		fs.srvProxy.DstList[DstIndex].Counter++
		if *configInfo.Debug {
			fmt.Printf("DstAddr:%s Remote:%s\n", fs.srvProxy.GetDstAddr(DstIndex), localConn.RemoteAddr().String())
		}
		go fs.Forward(localConn, DstIndex)
	    
	case "Weight":
		DstIndex := -1
		
		for k,Dst := range fs.srvProxy.DstList {			
			if Dst.WeightCounter % Dst.Weight != 0 {
				DstIndex = k
				fs.srvProxy.DstList[k].WeightCounter++				
				break				
			}
		}
		//fmt.Printf("First DstIndex:%d\n",DstIndex)		
		if DstIndex == -1 {
			
			for k,Dst := range fs.srvProxy.DstList {
				if Dst.WeightCounter % Dst.Weight == 0 && DstIndex == -1 {
					DstIndex = k					
				}/* else {
					fs.srvProxy.DstList[k].WeightCounter = 0
				}	*/
				fs.srvProxy.DstList[k].WeightCounter = 1
				
				//fmt.Printf("Name:%s Weight:%v WeightCounter:%v DstIndex:%d\n",Dst.Name,Dst.Weight,Dst.WeightCounter,DstIndex)
			}		
		}
		//fmt.Printf("Final DstIndex:%d\n",DstIndex)
		DstIndex = fs.GetHealthNode(DstIndex)
		fs.srvProxy.DstList[DstIndex].Counter++
		if *configInfo.Debug {
			fmt.Printf("DstAddr:%s Remote:%s\n", fs.srvProxy.GetDstAddr(DstIndex), localConn.RemoteAddr().String())
		}
		go fs.Forward(localConn, DstIndex)
		
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
			fmt.Printf("DstAddr:%s Remote:%s DstIndex:%d Client:%v\n", fs.srvProxy.GetDstAddr(client.DstIndex), localConn.RemoteAddr().String(), client.DstIndex, client)
		}

		go fs.Forward(localConn, client.DstIndex)
	case "RoundRobin":
		fs.srvProxy.Index = fs.srvProxy.Counter % fs.srvProxy.DstLen
		fs.srvProxy.Counter++
		DstIndex := fs.GetHealthNode(fs.srvProxy.Index)
		fs.srvProxy.DstList[DstIndex].Counter++
		if *configInfo.Debug {
			fmt.Printf("DstAddr:%s Remote:%s\n", fs.srvProxy.GetDstAddr(DstIndex), localConn.RemoteAddr().String())
		}
		go fs.Forward(localConn, DstIndex)
	}
}

func (fs *ForwardServer) Reload(srvProxy Proxy) {
	fs.srvProxy = srvProxy
	fmt.Printf("FS:%s reloaded.\n", fs.srvProxy.Name)

}

func (fs *ForwardServer) Stop() {
	fs.Run = false
	if fs.localListener != nil {
		fs.localListener.Close()
	}
}

func (fs *ForwardServer) Listen(srvProxy Proxy) {
	
	if *configInfo.Debug {
		fmt.Printf("ForwardServer connType:%s serverAddr:%s Type:%s\n", srvProxy.Mode, srvProxy.GetSrcAddr(), srvProxy.Type)
	}
		
	fs.ClientList = list.New()
	fs.Run = true
	fs.srvProxy = srvProxy
	fs.srvProxy.Counter = 0
	if fs.srvProxy.CheckTime == 0 {
		fs.srvProxy.CheckTime = 5
	}
	fs.srvProxy.DstLen = len(fs.srvProxy.DstList)
	if fs.srvProxy.Mode == "tcp" || fs.srvProxy.Mode == "http" || fs.srvProxy.Mode == "health" {
	
		//Init
		switch fs.srvProxy.Type {
		case "Weight":			
			for k,Dst := range fs.srvProxy.DstList {				
				fs.srvProxy.DstList[k].Weight += 2					
				fs.srvProxy.DstList[k].WeightCounter = 1
				fmt.Printf("Name:%s Weight:%v WeightCounter:%v \n",Dst.Name,fs.srvProxy.DstList[k].Weight,fs.srvProxy.DstList[k].WeightCounter)
			}
			
		}
		
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

			fmt.Printf("FS:%s is stoped.\n", fs.srvProxy.Name)
		}

	} else {
		fmt.Printf("Unsupport mode:%s,listen failed.\n", fs.srvProxy.Mode)
	}
}
