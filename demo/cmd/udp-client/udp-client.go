// package main implements a UDP client that sends UDP data to a UDP echo server and prints the response.
package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	name := os.Args[1]
	port := os.Args[2]

	nameport := name + ":" + port

	for {
		conn, err := net.Dial("udp", nameport)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Connected: %T, %v\n", conn, conn)

		fmt.Printf("Local address: %v\n", conn.LocalAddr())
		fmt.Printf("Remote address: %v\n", conn.RemoteAddr())

		b := []byte(os.Args[3])
		cc, wrerr := conn.Write(b)

		if wrerr != nil {
			fmt.Printf("conn.Write() error: %s\n", wrerr)
		} else {
			fmt.Println()
			fmt.Println(time.Now())
			fmt.Printf("Wrote %d bytes to socket\n", cc)
			c := make([]byte, cc+10)
			cc, rderr := conn.Read(c)
			if rderr != nil {
				fmt.Printf("conn.Read() error: %s\n", rderr)
			} else {
				fmt.Printf("Read %d bytes from socket\n", cc)
				fmt.Printf("Bytes: %q\n", string(c[0:cc]))
			}
		}
		if err = conn.Close(); err != nil {
			log.Fatal(err)
		}

		time.Sleep(time.Second * 2)
	}
}
