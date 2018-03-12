package main

import (
	"io"
	"log"
	"net"
	"time"
)

func recv(conn io.Reader, toTAP chan []byte) {
	var remaining []byte = make([]byte, 0)
	b := make([]byte, 10000)
	lenb := 0
	for {
		n, err := conn.Read(b)
		log.Printf("read %d\n", n)
		if err != nil {
			log.Println("Closing connection")
			return
		}
		s := 0
		for s < n {
			if lenb == 0 {
				lenb = bytesLen(b[s : s+2])
				s += 2
				continue
			}
			if s+lenb <= n {
				var packet []byte
				if len(remaining) == 0 {
					packet = b[s : s+lenb]
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
	b[0] = byte(length & 0xff)
	b[1] = byte((length >> 8) & 0xff00)
	return b
}

func bytesLen(b []byte) int {
	return int(b[0]) | (int(b[1]) << 8)
}

func doTCPServer(listenAddr string, toTAP chan []byte, fromTAP chan []byte) {
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

func doTCPClient(dest string, toTAP chan []byte, fromTAP chan []byte) {
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
