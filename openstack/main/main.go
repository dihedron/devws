package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/dihedron/devws/openstack"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Error("error loading .env file", "error", err)
	}

	slog.SetLogLoggerLevel(slog.LevelDebug)
	client, err := openstack.NewClient("")
	if err != nil {
		slog.Error("error creating client", "error", err)
		os.Exit(1)
	}
	servers, err := client.ComputeV2.List(context.TODO())
	if err != nil {
		slog.Error("error listing servers", "error", err)
		os.Exit(1)
	}
	for _, server := range servers {
		fmt.Printf("server (id: %s, name: %s)\n", server.ID, server.Name)
	}
}
