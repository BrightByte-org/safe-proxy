package main

import (
	"github.com/iooojik-dev/proxy/internal/proxy/http_proxy"
	"log"
	"net/http"
)

func main() {
	runHttpProxyServer()
}

func runHttpProxyServer() {
	server := &http_proxy.HttpProxy{Addr: "127.0.0.1:8080"}
	log.Fatal(http.ListenAndServe(server.Addr, server))
}
