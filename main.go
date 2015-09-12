package main

import (
	"flag"
	"log"
	"net/http"
	"text/template"
	"os"
)

var homeTempl = template.Must(template.ParseFiles("home.html"))

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl.Execute(w, r.Host)
}

func main() {
	flag.Parse()
	//go h.run()
	http.HandleFunc("/", serveHome)
	//http.HandleFunc("/ws", serveWs)
	err := http.ListenAndServe(":"+os.GetEnv("PORT"), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
