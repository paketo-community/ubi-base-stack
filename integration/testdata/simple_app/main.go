package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		fmt.Fprintln(w, runtime.Version())
	})

	http.HandleFunc("/nodejs/version", func(w http.ResponseWriter, r *http.Request) {

		cmd := exec.Command("node", "--version")

		output, err := cmd.Output()
		if err != nil {
			http.Error(w, "Error getting Node.js version", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(output)
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil))
}
