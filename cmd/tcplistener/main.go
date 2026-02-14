package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

// const filename = "messages.txt"
const port = 42069

func main() {

	// f, err := os.Open(filename)
	// if err != nil {
	// 	fmt.Println("error opening filename: ", filename)
	// 	fmt.Println("error: ", err)
	// }
	// ch := getLinesChannel(f)
	// for s := range ch {
	// 	fmt.Printf("read: %s\n", s)
	// }
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("unable to listen on port: %d, error: %v", port, err)
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Printf("unable to connection, error: %v", err)
			conn.Close()
		}

		fmt.Println("connection accepted lets go")
		ch := getLinesChannel(conn)
		for s := range ch {
			fmt.Printf("%s\n", s)
		}
	}
}

func getLinesChannel(r io.ReadCloser) <-chan string {
	ch := make(chan string)

	go func() {
		defer func() {
			defer r.Close()
			close(ch)
			fmt.Println("closing channel and reader")
		}()
		sb := strings.Builder{}
		for {
			buf := make([]byte, 8)
			n, err := r.Read(buf)
			if err != nil {
				if err == io.EOF {
					s := readLine(buf[:n], &sb)
					if s != nil {
						ch <- *s
					}

					remaining := sb.String()
					if len(remaining) > 0 {
						ch <- remaining
					}
					break
				}
				fmt.Println("error reading from file: ", err)
				os.Exit(1)
			}

			s := readLine(buf[:n], &sb)
			if s != nil {
				ch <- *s
			}
		}

	}()
	return ch
}

func readLine(b []byte, sb *strings.Builder) *string {
	if i := strings.Index(string(b), "\n"); i >= 0 {
		sb.Write(b[:i])
		s := sb.String()
		sb.Reset()

		if i+1 < len(b) {
			sb.Write(b[i+1:])
		}

		return &s
	}

	sb.Write(b)
	return nil
}
