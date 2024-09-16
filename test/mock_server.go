package main

import (
    "fmt"
    "net/http"
)

func main() {

	// First server on port 8080
    go func() {
        http.HandleFunc("/get-actual-host", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, `{"host": "menu.live"}`)
		})
        fmt.Println("Mock server running on http://localhost:5214")
        if err := http.ListenAndServe(":5214", nil); err != nil {
            fmt.Printf("/get-actual-host error: %v\n", err)
        }
    }()

    // Second server on port 8081
    go func() {
        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
            fmt.Fprintln(w, "This is upstream host")
        })
        fmt.Println("Upstream running on http://localhost:5215")
        if err := http.ListenAndServe(":5215", nil); err != nil {
            fmt.Printf("Upstream error: %v\n", err)
        }
    }()

    // Block the main goroutine to keep the servers running
    select {}
}