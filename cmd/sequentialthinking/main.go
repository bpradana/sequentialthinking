package main

import (
	"context"
	"flag"
	"log"

	"github.com/bpradana/sequentialthinking/internal/server"
	"github.com/bpradana/sequentialthinking/internal/thinking"
)

var httpAddr = flag.String("http", "", "if set, use streamable HTTP at this address, instead of stdin/stdout")

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	store := thinking.NewMemoryStore()
	srv := server.New(store)

	return server.Run(ctx, srv, *httpAddr)
}
