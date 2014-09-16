package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/sergi/go-diff/diffmatchpatch"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func serve(kill chan bool) {
	log.Println("Serving...")

	baseFormatBytes, err := ioutil.ReadFile("public/base.html")
	if nil != err {
		log.Fatal(err)
	}

	baseColumnFormatBytes, err := ioutil.ReadFile("public/baseColumns.html")
	if nil != err {
		log.Fatal(err)
	}

	baseDiffBytes, err := ioutil.ReadFile("public/baseDiff.html")
	if nil != err {
		log.Fatal(err)
	}

	baseFormat := string(baseFormatBytes)
	baseColumnFormat := string(baseColumnFormatBytes)
	baseDiff := string(baseDiffBytes)

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, baseFormat, "status...", "status...", status())
	})

	r.HandleFunc("/diff/{branch}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		branch := vars["branch"]
		d := diff(branch)
		fmt.Fprintf(w, baseFormat, branch, branch, d)
	})

	r.HandleFunc("/diff/stat/{branch}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		branch := vars["branch"]
		d := diffStat(branch)
		fmt.Fprintf(w, baseFormat, branch, branch, d)
	})

	r.HandleFunc("/show/{branch}/{filename:.*}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		b := vars["branch"]
		filename := vars["filename"]
		w.Header()["x-filename"] = []string{filename}
		fmt.Fprintf(w, baseFormat, filename, filename, gitReadFile(b, filename))
	})

	r.HandleFunc("/split-diff", func(w http.ResponseWriter, r *http.Request) {
		leftBranch := r.FormValue("leftBranch")
		rightBranch := r.FormValue("rightBranch")
		filename := r.FormValue("filename")

		title := fmt.Sprintf("%s..%s -- %s", leftBranch, rightBranch, filename)

		diff := diffmatchpatch.New()
		leftFile := gitReadFile(leftBranch, filename)
		rightFile := gitReadFile(rightBranch, filename)

		fmt.Fprintf(w,
			baseDiff,
			title,
			title,
			diff.DiffMain(leftFile, rightFile, true),
			leftFile,
			rightFile,
		)
	})

	r.HandleFunc("/split-diff.json", func(w http.ResponseWriter, r *http.Request) {
		leftBranch := r.FormValue("leftBranch")
		rightBranch := r.FormValue("rightBranch")
		filename := r.FormValue("filename")

		diff := diffmatchpatch.New()
		leftFile := gitReadFile(leftBranch, filename)
		rightFile := gitReadFile(rightBranch, filename)

		json, err := json.Marshal(diff.DiffMain(leftFile, rightFile, true))
		if nil != err {
			w.WriteHeader(http.StatusNotFound)
		}
		w.Header()["content-type"] = []string{"application/json"}
		w.Write(json)
	})

	r.HandleFunc("/split/", func(w http.ResponseWriter, r *http.Request) {
		leftBranch := r.FormValue("leftBranch")
		rightBranch := r.FormValue("rightBranch")
		filename := r.FormValue("filename")

		title := fmt.Sprintf("%s..%s -- %s", leftBranch, rightBranch, filename)

		leftFile := gitReadFile(leftBranch, filename)
		rightFile := gitReadFile(rightBranch, filename)

		fmt.Fprintf(w,
			baseColumnFormat,
			title,
			title,
			leftFile,
			rightFile,
		)
	})

	r.HandleFunc("/kill", func(w http.ResponseWriter, r *http.Request) {
		close(kill)
	})

	n := negroni.Classic()
	n.UseHandler(r)

	log.Println("Listening...")
	n.Run(":7777")
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

func diffStat(commit string) string {
	return git("diff", "--stat", "-M", commit)
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
