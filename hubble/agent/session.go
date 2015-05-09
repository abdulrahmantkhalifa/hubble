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

func StartLocalSession(conn *hubble.Connection, initiator *hubble.InitiatorMessage) {
	log.Printf("Starting session: (%v) %v:%v", initiator.GUID, initiator.Ip, initiator.Port)
	go func() {
		defer func(){
			//send terminator message.
			conn.Send(hubble.TERMINATOR_MESSAGE_TYPE, &hubble.TerminatorMessage {
				GuidMessage: hubble.GuidMessage{initiator.GUID},
			})
		}()

		//make local connection
		socket, err := net.Dial("tcp", fmt.Sprintf("%s:%d", initiator.Ip, initiator.Port))
		conn.SendAckOrError(initiator.GUID, err)
		if err != nil {
			return
		}

		defer socket.Close()

		channel := make(chan *hubble.MessageCapsule, sessionQueueSize)

		sessions[initiator.GUID] = channel

		ServeSession(initiator.GUID, conn, channel, socket)
	} ()
}

func ServeSession(guid string, conn *hubble.Connection, channel chan *hubble.MessageCapsule, socket net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		//socket -> proxy 
		defer func(){
			wg.Done()
			//force the receiver routine to exit
			channel <- &hubble.MessageCapsule{
				Mtype: hubble.INVALID_MESSAGE_TYPE,
			}
		}()
		
		buffer := make([]byte, 1024)
		for {
			count, read_err := socket.Read(buffer)
			if read_err != nil && read_err != io.EOF {
				//log.Printf("Failer on session %v %v: %v", guid, tunnel, read_err)
				return
			}

			err := conn.Send(hubble.DATA_MESSAGE_TYPE, &hubble.DataMessage{
				GuidMessage: hubble.GuidMessage{guid},
				Data: buffer[0:count],
			})

			if err != nil {
				//failed to forward data to proxy
				return
			}

			if read_err == io.EOF {
				return
			}
		}
	}()

	go func() {
		//proxy -> socket
		defer wg.Done()
		
		for {
			msgCap := <- channel
			//send on open socket.
			if msgCap.Mtype == hubble.INVALID_MESSAGE_TYPE || msgCap.Mtype == hubble.TERMINATOR_MESSAGE_TYPE {
				//force socket termination
				socket.Close()
				return
			}

			if msgCap.Mtype == hubble.DATA_MESSAGE_TYPE {
				data := msgCap.Message.(*hubble.DataMessage)
				written := 0
				for len(data.Data) != written {
					count, err := socket.Write(data.Data[written:])
					if err != nil {
						return
					}

					written += count
				}
			}
		}
	}()

	wg.Wait()
}