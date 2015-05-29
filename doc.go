// Package camillo is an idiomatic approach to web middleware in Go. It is tiny, non-intrusive, and encourages use of net/http Handlers.
//
// If you like the idea of Martini, but you think it contains too much magic, then Camillo is a great fit.
//
// For a full guide visit http://github.com/codegangsta/camillo
//
//  package main
//
//  import (
//    "github.com/codegangsta/camillo"
//    "net/http"
//    "fmt"
//  )
//
//  func main() {
//    mux := http.NewServeMux()
//    mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
//      fmt.Fprintf(w, "Welcome to the home page!")
//    })
//
//    n := camillo.Classic()
//    n.UseHandler(mux)
//    n.Run(":3000")
//  }
package camillo
