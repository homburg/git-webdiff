package main

import (
	"fmt"
	// "github.com/sergi/go-diff/diffmatchpatch"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func serve(kill chan bool) {
	log.Println("Serving...")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header()["content-type"] = []string{"text/plain"}
		fmt.Fprint(w, status())
		fmt.Fprintln(w)
		fmt.Fprintln(w, "==================")
		fmt.Fprintln(w, gitReadFile("develop", "README.md"))
	})

	http.HandleFunc("/diff/", func(w http.ResponseWriter, r *http.Request) {
		w.Header()["content-type"] = []string{"text/plain"}
		d := diff("develop")
		fmt.Fprint(w, d)
	})

	http.HandleFunc("/kill", func(w http.ResponseWriter, r *http.Request) {
		close(kill)
	})

	log.Println("Listening...")
	err := http.ListenAndServe(":7777", nil)
	if nil != err {
		log.Fatal(err)
	}
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

func gitCurrentBranch() string {
	branchInfo := git("branch")

	lines := strings.Split(branchInfo, "\n")
	for _, line := range lines {
		if line[0] == '*' {
			return line[2:]
		}
	}
	return ""
}

func gitReadFile(branch, filename string) string {
	if branch == "" {
		branch = gitCurrentBranch()
	}

	return git("show", fmt.Sprintf("%s:%s", branch, filename))
}

func main() {
	log.Println("Starting...")
	kill := make(chan bool)

	if len(os.Args) == 1 {
		log.Println("No target...")
	}

	commit := os.Args[1]

	go serve(kill)
	go func() {
		addr := fmt.Sprintf("http://localhost:7777/diff/%s", commit)
		log.Printf(`Opening "%s"`+"\n", addr)
		cmd := exec.Command("xdg-open", addr)
		cmd.Start()
	}()

	for _ = range kill {
	}
	log.Println("Ending...")
}
