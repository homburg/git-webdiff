package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func serve(kill chan bool) {
	log.Println("Serving...")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header()["content-type"] = []string{"text/plain"}
		fmt.Fprint(w, status())
	})

	http.HandleFunc("/diff/", func(w http.ResponseWriter, r *http.Request) {
		w.Header()["content-type"] = []string{"text/plain"}
		d := diff("develop")
		fmt.Fprint(w, d)
	})

	http.HandleFunc("/kill", func(w http.ResponseWriter, r *http.Request) {
		close(kill)
	})

	http.ListenAndServe(":7777", nil)
}

func git(args ...string) string {
	cmd := exec.Command("git", args...)

	output, err := cmd.CombinedOutput()
	if nil != err {
		return err.Error()
	}

	return string(output)
}

func status() string {
	return git("status")
}

func diff(commit string) string {
	return git("diff", "-M", commit)
}

func main() {
	kill := make(chan bool)

	if len(os.Args) == 1 {
		log.Println("No target...")
	}

	commit := os.Args[1]

	go serve(kill)
	go func() {
		addr := fmt.Sprintf("http://localhost:7777/%s", commit)
		log.Printf(`Opening "%s"`+"\n", addr)
		cmd := exec.Command("xdg-open", addr)
		cmd.Start()
	}()

	for _ = range kill {
	}
	log.Println("Ending...")
}
