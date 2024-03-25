package main

import (
	"flag"
	"fmt"
	"imgProcessing/api"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	addr := flag.String("addr", ":5000", "server port")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{\"message\": \"hello world\"}"))
	}).Methods("GET")

	r.HandleFunc("/convert", api.Convert).Methods(http.MethodPost)
	r.HandleFunc("/resize", api.Resize).Methods(http.MethodPost)
	r.HandleFunc("/compress", api.Compress).Methods(http.MethodPost)

	fmt.Printf("server run on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, r))
}
