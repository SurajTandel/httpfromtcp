package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal("error resolving UDP address:", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal("error dialing UDP:", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Println("error reading from stdin:", err)
			continue
		}

		// Remove the newline character before sending
		line = strings.TrimRight(line, "\n\r")
		
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Println("error writing to UDP connection:", err)
			continue
		}
	}
}

