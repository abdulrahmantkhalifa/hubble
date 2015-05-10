package main 

import (
	"hubble/proxy"
	"net/http"
	"log"
)



func main() {
	http.HandleFunc("/", proxy.ProxyHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
