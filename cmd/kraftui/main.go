package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ijaidev/kraftui/internal/server"
)

// version can be overridden at build time:
//
//	go build -ldflags "-X main.version=0.1.0" ./cmd/kraftui
var version = "dev"

func main() {
	port := flag.Int("port", 0, "HTTP listen port")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := server.NewServer(*port, version)

	err := server.Listen(ctx)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
