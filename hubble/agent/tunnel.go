package agent

import (
	// "hubble"
	// "net"
)

// func OpenTunnel(tunnel *hubble.Tunnel) error {
// 	// open socket and wait for connections
// 	listner, err := net.Listen("tcp", fmt.Sprintf(":%d", tunnel.Local))
// 	defer listner.Close()

// 	if err != nil {
// 		return err
// 	}

// 	for {
// 		conn, err := listner.Accept()
// 		if err != nil {
// 			//log.Printf("accept socket error: %v", err)
// 			conn.Close()
// 		}

// 		// go handle(conn)
// 	}
// }