package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/songgao/water"
)

func recv(conn0 io.Reader, toTAP chan []byte) {
	var remaining []byte = make([]byte, 0)
	b := make([]byte, 10000)
	lenb := 0
	conn := bufio.NewReader(conn0)
	for {
		n, err := conn.Read(b)
		if err != nil {
			log.Println("Closing connection")
			return
		}
		s := 0
		for s < n {
			if lenb == 0 {
				lenb = bytesLen(b[s:s+2])
				s += 2
				continue
			}
			if s + lenb <= n {
				var packet []byte
				if len(remaining) == 0 {
					packet = b[s:s+lenb]
				} else {
					packet = append(remaining, b[s:s+lenb]...)
					remaining = make([]byte, 0)
				}
				s += lenb
				lenb = 0
				toTAP <- packet
			} else {
				remaining = append(remaining, b[s:n]...)
				lenb -= n - s
				s = n
			}
		}
	}
}

func lenBytes(length int) []byte {
	b := make([]byte, 2)
	b[0] = (byte)(length % 256)
	b[1] = (byte)((length / 256) % 256)
	return b
}

func bytesLen(b []byte) int {
	return int(b[0]) + int(b[1]) * 256
}

func doServer(listenAddr string, toTAP chan []byte, fromTAP chan []byte) {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		panic("Listen failed")
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			panic("Accept failed")
		}
		defer conn.Close()
		go encodeAndSend(conn, fromTAP)
		recv(conn, toTAP)
	}
}

func doClient(dest string, toTAP chan []byte, fromTAP chan []byte) {
	for {
		raddr, err := net.ResolveTCPAddr("tcp", dest)
		conn, err := net.DialTCP("tcp", nil, raddr)
		if err != nil {
			log.Println("Dial failed")
			time.Sleep(3 * time.Second)
			continue
		}
		defer conn.Close()
		log.Println("connected")
		go encodeAndSend(conn, fromTAP)
		recv(conn, toTAP)
	}
}

func encodeAndSend(conn io.Writer, fromTAP chan []byte) {
	for b := range fromTAP {
		conn.Write(lenBytes(len(b)))
		conn.Write(b)
	}
}

func startTAP(tapName string, persist bool) *water.Interface {
	config := water.Config{
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name: tapName,
			Persist: persist,
		},
		DeviceType: water.TAP,
	}
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
			//panic(err)
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
	tapName := os.Args[1]
	destAddr := os.Args[2]
	listenAddr := os.Args[3]

	toTAP := make(chan []byte)
	fromTAP := make(chan []byte)
	ifce := startTAP(tapName, true)
	go sendToTAP(ifce, toTAP)
	go readFromTAP(ifce, fromTAP)
	if listenAddr != "" {
		doServer(listenAddr, toTAP, fromTAP)
	} else {
		doClient(destAddr, toTAP, fromTAP)
	}
}
