package main

import (
	"flag"
	"io/ioutil"
	"log"
	"sync"

	"github.com/devansh42/shree2/serv/relay"
)

func main() {

	caCertFilePath := flag.String("caCert", "/tmp", "CA Root certificate filepath")
	caPrivFilePath := flag.String("caPriv", "/tmp", "CA private key filepath")
	relayCertFilePath := flag.String("relayCert", "/tmp", "Relay Server certificate filepath")
	flag.Parse()
	for !flag.Parsed() {
	}

	caCert := getFileContent(*caCertFilePath)
	caPriv := getFileContent(*caPrivFilePath)
	relayCert := getFileContent(*relayCertFilePath)

	rel, err := relay.New(caCert, caPriv, relayCert)
	if err != nil {
		log.Fatal("Couldn't initialize relay: ", err)
	}
	wg := new(sync.WaitGroup)
	wg.Add(2)

	go rel.StartApplicationServer()
	go rel.StartRelayServer()

	wg.Wait()
}

func getFileContent(filePath string) []byte {
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("Couldn't load file content: ", err)
	}
	return fileContent
}
