package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ijaidev/kraftui/config"
	"github.com/ijaidev/kraftui/internal/server"
	"github.com/ijaidev/kraftui/log"
)

func main() {
	options, err := loadCLI(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(2)
	}
	if options.help {
		writeHelp(os.Stdout, options.parser)
		return
	}
	if options.values.Version {
		fmt.Println(config.KraftVersion())
		return
	}
	config.Load(options.values.Port, options.values.LogType, options.values.LogLevel, options.values.SuppressLogs)
	log.Configure()
	log.G.Info("logger configured", "type", config.CurrentLogType(), "level", config.CurrentLogLevel(), "suppressed", config.SuppressLogs())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := server.NewServer(config.Port(), config.KraftVersion())

	err = server.Listen(ctx)

	if err != nil {
		log.G.Error("server stopped", "error", err, "report_issue", "https://github.com/ijaidev/kraftui")
		os.Exit(1)
	}
}
