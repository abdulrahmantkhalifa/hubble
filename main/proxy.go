package main

import (
	"flag"
	"fmt"
	"github.com/Jumpscale/hubble/proxy"
	"log"
	"net/http"
)

func main() {
	var listenAddr string
	var help bool

	flag.BoolVar(&help, "h", false, "Print this help screen")
	flag.StringVar(&listenAddr, "listen", ":8080", "Listining address")
	flag.Parse()

	printHelp := func() {
		fmt.Println("proxy [options]")
		flag.PrintDefaults()
	}

	if help {
		printHelp()
		return
	}

	log.Println("Start listing on", listenAddr)
	http.HandleFunc("/", proxy.ProxyHandler)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
