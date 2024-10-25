package main

import (
    "log"
    "github.com/vishal/snow/internal/server"
)

func main() {
    srv := server.New()
    log.Fatal(srv.Start(":8080"))
}