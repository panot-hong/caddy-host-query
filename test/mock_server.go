package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/get-actual-host", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, `{"host": "google.com"}`)
    })

    fmt.Println("Mock server running on http://localhost:5214")
    http.ListenAndServe(":5214", nil)
}