package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mattesthaut/gdfs/core"
)

func main() {
	port := flag.Int("port", 42042, "DFS node port")
	bootstrapAddr := flag.String("bootstrap", "", "Bootstrap address")
	flag.Parse()

	storage := core.NewMemoryStorage()

	nodeAddr := fmt.Sprintf("127.0.0.1:%d", *port)
	host := core.NewHost(nodeAddr, storage)

	if err := host.Start(); err != nil {
		log.Fatal(err)
	}

	if *bootstrapAddr != "" {
		if err := host.Bootstrap(*bootstrapAddr); err != nil {
			log.Fatal(err)
		}
	}

	go func() {
		for {
			time.Sleep(60 * time.Second)
			peerCount := host.KnownPeerCount()
			log.Printf("%d node in routing table", peerCount)
			log.Printf("%d value in storage", storage.Len())
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	host.Stop()
}
