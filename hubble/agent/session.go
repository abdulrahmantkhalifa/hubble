package agent

import (
	"hubble"
	"net"
	"fmt"
	"sync"
	"io"
	"log"
)

const sessionQueueSize = 512
var sessions = make(map[string]chan *hubble.MessageCapsule)

func startLocalSession(conn *hubble.Connection, initiator *hubble.InitiatorMessage) {
	log.Printf("Starting local session: (%v) %v:%v", initiator.GUID, initiator.Ip, initiator.Port)
	go func() {
		// defer func(){
		// 	//send terminator message.
		// 	conn.Send(hubble.TERMINATOR_MESSAGE_TYPE, &hubble.TerminatorMessage {
		// 		GuidMessage: hubble.GuidMessage{initiator.GUID},
		// 	})
		// }()

		//make local connection
		socket, err := net.Dial("tcp", fmt.Sprintf("%s:%d", initiator.Ip, initiator.Port))
		conn.SendAckOrError(initiator.GUID, err)
		if err != nil {
			log.Println(err)
			return
		}

		defer socket.Close()

		//channel := make(chan *hubble.MessageCapsule, sessionQueueSize)
		channel := make(chan *hubble.MessageCapsule) //not buffered for testing
		defer close(channel)

		sessions[initiator.GUID] = channel
		defer delete(sessions, initiator.GUID)
		
		serveSession(initiator.GUID, conn, channel, socket)
	} ()
}

func serveSession(guid string, conn *hubble.Connection, channel chan *hubble.MessageCapsule, socket net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		//socket -> proxy 

		defer func() {
			wg.Done()
			//send teminator to local channel (in case it still waiting)
			defer func() {
				recover()
			} ()
			
			conn.Send(hubble.CONNECTION_CLOSED_MESSAGE_TYPE,
					  &hubble.ConnectionClosedMessage{
					  	 GuidMessage: hubble.GuidMessage{guid},
					  })

			//force closing the local receiver
			channel <- &hubble.MessageCapsule{
				Mtype: hubble.TERMINATOR_MESSAGE_TYPE,
			}
		} ()
		
		buffer := make([]byte, 1024)
		order := 1

		for {
			count, read_err := socket.Read(buffer)
			if read_err != nil && read_err != io.EOF {
				//log.Printf("Failer on session %v %v: %v", guid, tunnel, read_err)
				return
			}

			err := conn.Send(hubble.DATA_MESSAGE_TYPE, &hubble.DataMessage{
				GuidMessage: hubble.GuidMessage{guid},
				Order: order,
				Data: buffer[:count],
			})

			order ++

			if err != nil {
				//failed to forward data to proxy
				log.Println(err)
				return
			}

			if read_err == io.EOF {
				return
			}
		}
	}()

	go func() {
		//proxy -> socket

		defer func() {
			wg.Done()
		}()

		lastOrder := 0
		for {
			msgCap, ok := <- channel
			//send on open socket.
			if !ok || msgCap.Mtype == hubble.TERMINATOR_MESSAGE_TYPE {
				//force socket termination
				socket.Close()
				return
			}

			if msgCap.Mtype == hubble.DATA_MESSAGE_TYPE {
				data := msgCap.Message.(*hubble.DataMessage)
				if lastOrder + 1 != data.Order {
					log.Println("Data out of order")
					socket.Close()
					return
				}

				lastOrder = data.Order

				written := 0
				for written < len(data.Data) {
					count, err := socket.Write(data.Data[written:])
					if err != nil {
						log.Println(err)
						return
					}

					written += count
				}
			}
		}
	}()

	wg.Wait()
	log.Printf("Session '%v' routines terminates", guid)
}