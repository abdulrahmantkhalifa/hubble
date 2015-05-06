package main 

import (
	"hubble/proxy"
	"net/http"
	"log"
)

// var upgrader = websocket.Upgrader {
//     ReadBufferSize:  1024,
//     WriteBufferSize: 1024,
//     CheckOrigin: func(request *http.Request) bool { return true },
// }

// func handler(writer http.ResponseWriter, request *http.Request) {
// 	conn, err := upgrader.Upgrade(writer, request, nil)
// 	defer conn.Close()

//     if err != nil {
//         log.Println(err)
//         return
//     }
	
// 	for {
// 		mode, reader, err := conn.NextReader()
// 		if err != nil {
// 			log.Println(err)
// 			break
// 		}

// 		if mode != websocket.BinaryMessage {
// 			//only binary messages are supported.
// 			log.Println("Only binary messages are supported")
// 			break
// 		}
		
// 		mtype, message, err := hubble.Loads(reader)
// 		log.Println(mtype, message, err)
// 	}

// 	log.Println("Done reading")
// }


func main() {
	http.HandleFunc("/", proxy.ProxyHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}