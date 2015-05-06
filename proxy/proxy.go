package main 

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader {
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(request *http.Request) bool { return true },
}

func handler(writer http.ResponseWriter, request *http.Request) {
	log.Println("Received connection")
	conn, err := upgrader.Upgrade(writer, request, nil)
    if err != nil {
        log.Println(err)
        return
    }
	
	var buffer []byte = make([]byte, upgrader.ReadBufferSize)
	log.Println("Start reading")
	
	for {
		mtype, reader, error := conn.NextReader()
		log.Println("NReader", mtype, error)
		if error != nil {
			break
		}
		
		for {
			length, read_error := reader.Read(buffer)
			log.Println("Read", length, read_error)
			if read_error != nil {
				break
			}
			log.Println("Value:", string(buffer[:length]))
		}
	}

	log.Println("Done reading")
}


func main() {
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}