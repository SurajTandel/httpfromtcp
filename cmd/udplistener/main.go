package main

import (
	"log"
	"net"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":42069")
	if err != nil {
		log.Fatal("error resolving UDP address:", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal("error listening on UDP:", err)
	}
	defer conn.Close()

	log.Println("UDP listener started on :42069")

	buffer := make([]byte, 1024)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println("error reading from UDP:", err)
			continue
		}

		message := string(buffer[:n])
		log.Printf("Received from %s: %s\n", clientAddr, message)
	}
}

