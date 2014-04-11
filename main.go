package main

import (
	"fmt"
	"net/http"
)

var disp Dispatcher

func main() {
	disp, err := NewDispatcher("rss.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("starting dispatcher")
	disp.Start()

	fmt.Println("starting http server")
	http.HandleFunc("/", ViewHandler)
	http.HandleFunc("/add/", AddHandler)
	http.HandleFunc("/remove/", RemoveHandler)
	fmt.Println(http.ListenAndServe(":8080", nil))
}
