package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"go-worker/config"
	"go-worker/data_adapters"
	"go-worker/logger"
	"go-worker/workerpool"
)

func main() {

	// capture os signals
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

	// Initialize config, logger, adapters
	environment := flag.String("e", config.ProdEnvironment, "specify the app environment")
	flag.Parse()
	config.Init(*environment)
	logger.Init()
	dataAdapters.Init()

	// Start worker pool
	pool, err := workerpool.New(config.GetConfig().GetInt("worker.count"))
	if err != nil {
		logger.Log.Fatalf("Worked pool initiation failed with error: %s", err.Error())
	}

	// Start pprof apis
	go func() {
		if err := http.ListenAndServe("localhost:6000", nil); err != nil {
			logger.Log.Fatalf("pprof server init failed with error: %s", err.Error())
		}
	}()

	<-signalChan
	// Stop worker pool
	pool.Close()
	logger.Log.Infoln("Successfully terminated worker pool")
}
