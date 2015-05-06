package pipe

import (
	"log"
	"net"
	"io"
	"sync"
)

const BUFFER_SIZE = 1024


func _readto(wg *sync.WaitGroup, from, to net.Conn) {
	defer wg.Done()

	var buffer []byte = make([]byte, BUFFER_SIZE, BUFFER_SIZE)

	for {
		var read_length, read_error = from.Read(buffer)
		if read_error != nil && read_error != io.EOF {
			log.Printf("Failed to read into buffer:%v\n", read_error)
			return
		}
		var write_length = 0
		for write_length < read_length {
			var l, err = to.Write(buffer[write_length:read_length])
			if err != nil {
				log.Printf("Failed to write buffer:%v\n", err)
				return
			}
			write_length += l
		}

		if read_error == io.EOF {
			return
		}
	}
}

//Pipes two net.Conn together. The method will close both sockets if connection on one end is terminated or an error is occuerd.
func Pipe(c1, c2 net.Conn) {
	//var ch chan []byte = make(chan []byte)
	defer func () {
		c1.Close()
		c2.Close()
	}()

	var wg sync.WaitGroup

	wg.Add(2)
	go _readto(&wg, c1, c2)
	go _readto(&wg, c2, c1)

	wg.Wait()
}