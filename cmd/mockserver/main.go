package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/dialecticanet-com/rmq-rpc-mockserver/app"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/config"
)

var (
	service    = "amqp-mockserver"
	version    = "dev"
	commitHash = "0000000"
	buildDate  = time.Now().String()
	envFile    = flag.String("env", "", "Path to the environment file")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	serviceInfo := config.ServiceInfo{
		Service:    service,
		Version:    version,
		CommitHash: commitHash,
		BuildDate:  buildDate,
	}

	cfg, err := config.NewConfig(*envFile, serviceInfo)
	if err != nil {
		fmt.Println("error loading config: ", err.Error())
		os.Exit(1)
	}

	if err := app.Run(ctx, cfg); err != nil {
		fmt.Println("error running mockserver:" + err.Error())
		os.Exit(1)
	}
}
