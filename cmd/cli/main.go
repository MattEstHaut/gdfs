package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/mattesthaut/gdfs/core"
	"github.com/mattesthaut/gdfs/data"
)

func main() {
	isStoreReq := flag.Bool("store", false, "Store a file")
	isFindReq := flag.Bool("find", false, "Find a file")
	nodeAddr := flag.String("addr", "127.0.0.1:42042", "Node address")
	file := flag.String("file", "", "Filepath")
	fileId := flag.String("id", "", "File id")
	flag.Parse()

	if (*isStoreReq && *isFindReq) || !(*isStoreReq || *isFindReq) {
		log.Fatal("Bad command")
	}

	storage := core.NewFakeStorage()

	host := core.NewHost("", storage)

	if err := host.Bootstrap(*nodeAddr); err != nil {
		log.Fatal(err)
	}

	if *isStoreReq {
		filePath, err := filepath.Abs(*file)
		if err != nil {
			log.Fatal(err)
		}

		file, err := data.ReadFile(filePath)
		if err != nil {
			log.Fatal(err)
		}

		id, replicaCount := data.StoreData(file, host)
		fmt.Printf("%s  (%d replicas)", id, replicaCount)
	} else {

		id, err := core.IdFromString(*fileId)
		if err != nil {
			log.Fatal(err)
		}

		filePath, err := filepath.Abs(*file)
		if err != nil {
			log.Fatal(err)
		}

		fileData, found := data.FindData(id, host)
		if !found {
			log.Fatal("File not found")
		}

		if err := data.WriteFile(filePath, fileData); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%d bytes written to %s", len(fileData), *file)
	}
}
