package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

const address = ":42069"

func main() {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Printf("unable to resolve udp addr error: %v\n", err)
		os.Exit(1)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Printf("unable to dial udp addr error: %v\n", err)
		os.Exit(1)
	}

	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("error reading str from reader ostdin: %v\n", err)
			continue
		}

		_, err = conn.Write([]byte(str))
		if err != nil {
			fmt.Printf("error writing to udp conn: %v\n", err)
		}
	}
}
