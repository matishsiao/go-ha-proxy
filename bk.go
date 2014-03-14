package main
		/*
		//Forward Server
		if srvProxy.Mode == "udp"  {
		 	laddr, err := net.ResolveUDPAddr("udp",localAddrString)
		 	if err != nil {
		 	 	fmt.Println("Resolv..")
		 	} 
		 	localListener, erl := net.ListenUDP("udp", laddr);  
		 	for {
		        // Setup localConn (type net.Conn)
		        localConn, err := localListener.Accept()
		        if err != nil {
		            log.Fatalf("listen.Accept failed: %v", err)
		        }
		        go forward(localConn,connType,serverAddrString)
		    }	
		    fmt.Printf("Unsupport UDP.\n")
		} else {
		*/	
		
		/*
			func (fs *ForwardServer) getSrvConn(connType string,serverAddrString string) *list.Element {
		add := true
		//fmt.Printf("FS.CP[%d] \n",FS.ConnectionPool.Len())
		//if FS.ConnectionPool.Len() > 0 {
			for e := fs.ConnectionPool.Front(); e != nil; e = e.Next() {
				e.Value.(*FSConnection).Conn.SetReadDeadline(time.Now())
				one := make([]byte,0);
				if _, err := e.Value.(*FSConnection).Conn.Read(one); err == io.EOF {
				  fmt.Printf("detected closed LAN connection:%d\n",e.Value.(*FSConnection).Id)
				  e.Value.(*FSConnection).Conn.Close()
				  //e.Value.(*FSConnection).Conn = nil
				  FS.ConnectionPool.Remove(e)  
				} else {
				  //var zero time.Time
				  e.Value.(*FSConnection).Conn.SetReadDeadline(time.Time{})
				  if e.Value.(*FSConnection).Status == false {
					e.Value.(*FSConnection).Status = true		
					return e
		          }
				}
				if e.Value.(*FSConnection).Status == false {
					e.Value.(*FSConnection).Status = true		
					return e
		        }				
			}	
		//}
		if add {
			conn := new(FSConnection)			
			srvConn, err := net.Dial(connType, serverAddrString)
			if err != nil {
				fmt.Printf("getSrvConn Err: %v\n", err)
				return nil
			}
			conn.Conn = srvConn
			conn.Status = true
			conn.Id = count
			e := fs.ConnectionPool.PushBack(conn)
			return e
		}
		return nil
	}
		*/