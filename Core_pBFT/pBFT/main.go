package main

import (
	"log"
	"os"
	"runtime"
)

func main() {

	logFile, err := os.OpenFile("logfile.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	log.Println("server on")

	runtime.GOMAXPROCS(runtime.NumCPU())
	nodeID := os.Args[1]
	server := NewServer(nodeID)

	server.Start()

}
