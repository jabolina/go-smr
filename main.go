package main

import (
	"flag"
	"log"
	"net/http"
	"smr/web"
)

var (
	httpAddr = flag.String("addr", "127.0.0.1:8080", "HTTP address")
	partitionName = flag.String("name", "", "Partition name")
	confPath = flag.String("conf", "./replicas.json", "Configuration file path")
)

func parser() {
	flag.Parse()
}

func main() {
	log.Println("Starting smr application")
	parser()

	srv, err := web.NewServer(*confPath, *partitionName)
	if err != nil {
		log.Fatalf("failed creating server %s: %v", *httpAddr, err)
	}
	http.HandleFunc("/get", srv.GetRequest)
	http.HandleFunc("/set", srv.SetRequest)
	log.Printf("start listening on [%s]", *httpAddr)
	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}
