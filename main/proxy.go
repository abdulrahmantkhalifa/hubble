package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/Jumpscale/hubble/auth"
	"github.com/Jumpscale/hubble/proxy"
)

func main() {
	var listenAddr string
	var help bool
	var authLua string
	var authAcceptAll bool

	flag.BoolVar(&help, "h", false, "Print this help screen")
	flag.StringVar(&listenAddr, "listen", ":8080", "Listining address")
	flag.StringVar(&authLua, "authlua", "", "Lua authorization module")
	flag.BoolVar(&authAcceptAll, "authall", false, "Grant all authorization requests")
	flag.Parse()

	printHelp := func() {
		fmt.Println("proxy [options]")
		flag.PrintDefaults()
	}

	if help {
		printHelp()
		return
	}

	if authLua != "" {
		// TODO: Initialize Lua authorization module.
		panic("Not implemented")
	} else {
		if !authAcceptAll {
			log.Println("Warning, no authorization module specified, will",
				"grant all authorization requests")
		}
		auth.Install(auth.NewAcceptAllModule())
	}
	log.Println("Start listing on", listenAddr)
	http.HandleFunc("/", proxy.ProxyHandler)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
