package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Jumpscale/hubble/auth"
	"github.com/Jumpscale/hubble/proxy"
)

func main() {
	var listenAddr string
	var help bool
	var authLua string
	var authAcceptAll bool
	var authDeclineAll bool

	flag.BoolVar(&help, "h", false, "Print this help screen")
	flag.StringVar(&listenAddr, "listen", ":8080", "Listining address")
	flag.StringVar(&authLua, "authlua", "", "Lua authorization module")
	flag.BoolVar(&authAcceptAll, "authgrant", false, "Grant all authorization requests")
	flag.BoolVar(&authDeclineAll, "authdeny", false, "Decline all authorization requests (for debugging purposes)")
	flag.Parse()

	printHelp := func() {
		fmt.Println("proxy [options]")
		flag.PrintDefaults()
	}

	if help {
		printHelp()
		return
	}

	if authDeclineAll {
		auth.Install(auth.NewDeclineAllModule())

	} else if authLua != "" {
		m, err := auth.NewLuaModule(authLua)
		if err != nil {
			log.Println("Cannot install Lua authorization module:", err)
			os.Exit(1)
		}

		auth.Install(m)

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
