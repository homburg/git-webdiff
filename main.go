package main

import (
	"bytes"
	"log"
	"net/http"
	"os/exec"
)

func serve(kill chan struct{}) {
	log.Println("Serving...")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header()["content-type"] = "text/plain"
		w.Write(status())
	})

	http.HandleFunc("/kill", func(w http.ResponseWriter, r *http.Request) {
		kill <- nil
	})

	http.ListenAndServe(":7777")
}

func status() string {
	cmd := exec.Command("git", "status")

	output, err := cmd.CombinedOutput()
	if nil != err {
		return err.Error()
	}

	return string(output)
}

func main() {
	log.Println("Starting...")
	kill := make(chan struct{})

	go serve(kill)

	_ <- kill
	log.Println("Ending...")
}
