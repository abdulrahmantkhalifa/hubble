package agent

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/Jumpscale/hubble"
)

type sessionChannel chan hubble.Message
type sessionsStore map[string]sessionChannel

func registerSession(sessions sessionsStore, guid string) sessionChannel {
	channel := make(sessionChannel)
	sessions[guid] = channel
	return channel
}

func unregisterSession(sessions sessionsStore, guid string) {
	channel, ok := sessions[guid]
	if ok {
		delete(sessions, guid)
		close(channel)
	}
}

func startLocalSession(sessions sessionsStore, conn *hubble.Connection, initiator *hubble.InitiatorMessage) {
	log.Printf("Starting local session: (%v) %v:%v", initiator.GUID, initiator.RemoteHost, initiator.RemotePort)
	go func() {
		//make local connection
		socket, err := net.Dial("tcp", fmt.Sprintf("%s:%d", initiator.RemoteHost, initiator.RemotePort))
		conn.SendAckOrError(initiator.GUID, err)
		if err != nil {
			log.Println(err)
			return
		}

		defer socket.Close()

		channel := registerSession(sessions, initiator.GUID)
		defer unregisterSession(sessions, initiator.GUID)

		serveSession(initiator.GUID, conn, channel, socket, nil)
	}()
}

func serveSession(guid string, conn *hubble.Connection, channel sessionChannel, socket net.Conn, ctrl ctrlChan) {
	log.Println("Starting routines for session", guid)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		//socket -> proxy

		defer func() {
			wg.Done()
			//send teminator to local channel (in case it still waiting)
			defer func() {
				recover()
			}()

			conn.Send(hubble.NewConnectionClosedMessage(guid))

			//force closing the local receiver
			channel <- hubble.NewTerminatorMessage(guid)
		}()

		buffer := make([]byte, 1024)
		var order int64 = 1

		for {
			count, read_err := socket.Read(buffer)
			if read_err != nil && read_err != io.EOF {
				//log.Printf("Failer on session %v %v: %v", guid, tunnel, read_err)
				return
			}

			err := conn.Send(hubble.NewDataMessage(guid, order, buffer[:count]))
			order++

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

		var lastOrder int64 = 0
		for {
			select {
			case message, ok := <-channel:
				//send on open socket.
				if !ok || message.GetMessageType() == hubble.TERMINATOR_MESSAGE_TYPE {
					//force socket termination
					socket.Close()
					return
				}

				if message.GetMessageType() == hubble.DATA_MESSAGE_TYPE {
					data := message.(*hubble.DataMessage)
					if lastOrder+1 != data.Order {
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
			case <-ctrl:
				socket.Close()
				return
			}
		}
	}()

	wg.Wait()
	log.Printf("Session '%v' routines terminates", guid)
}
