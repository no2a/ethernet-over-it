package main

import (
	"log"
	"os"

	"github.com/songgao/water"
)

func startTAP(tapName string, persist bool) *water.Interface {
	config := water.Config{
		DeviceType: water.TAP,
	}
        config.Name = tapName
        config.Persist = persist
	ifce, err := water.New(config)
	if err != nil {
		panic(err)
	}
	return ifce
}

func sendToTAP(ifce *water.Interface, C chan []byte) {
	for frame := range C {
		_, err := ifce.Write(frame)
		if err != nil {
			log.Printf("Write failed with %d bytes frame: %s\n", len(frame), err)
		}
	}
}

func readFromTAP(ifce *water.Interface, C chan []byte) {
	mtu := 65535
	frame := make([]byte, mtu)
	for {
		n, err := ifce.Read(frame)
		if err != nil {
			panic(err)
		}
		C <- frame[:n]
	}
}

func main() {
	toTAP := make(chan []byte)
	fromTAP := make(chan []byte)

	tapName := os.Args[1]
	kind := os.Args[2]
	if kind == "tcpserver" {
		listenAddr := os.Args[3]
		go doTCPServer(listenAddr, toTAP, fromTAP)
	} else if kind == "tcpclient" {
		destAddr := os.Args[3]
		go doTCPClient(destAddr, toTAP, fromTAP)
	} else if kind == "slack" {
		channel := os.Args[3]
		token := os.Getenv("TOKEN")
		go doSlackClient(token, channel, toTAP, fromTAP)
	} else {
		panic(kind)
	}
	ifce := startTAP(tapName, true)
	go sendToTAP(ifce, toTAP)
	readFromTAP(ifce, fromTAP)
}
