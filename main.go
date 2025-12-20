package main

import (
	"fmt"
	"net/http"
)

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<h1>Welcome to my site!</h1>")
}
func main() {
	const PORT string = ":3000"
	http.HandleFunc("/", handlerFunc)
	fmt.Println("Starting the server on", PORT)
	http.ListenAndServe(PORT, nil)
}
