package main

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"

	"github.com/instill-ai/connector-ai/config"
)

func main() {

	if err := config.Init(); err != nil {
		log.Fatal(err.Error())
	}
	ctx, cancel := context.WithCancel(context.Background())
	ctx, span := otel.Tracer("init-tracer").Start(ctx,
		"main",
	)
	defer span.End()
	defer cancel()
}
