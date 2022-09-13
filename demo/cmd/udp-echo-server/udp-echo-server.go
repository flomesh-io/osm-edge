// package main implements a UDP echo server that echoes back the UDP client's request as a part of its response.
package main

import (
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/openservicemesh/osm/pkg/logger"
)

var (
	log      = logger.NewPretty("udp-echo-server")
	logLevel = flag.String("logLevel", "debug", "Log output level")
	port     = flag.Int("port", 6060, "port on which this app is serving UDP connections")
)

func main() {
	flag.Parse()
	err := logger.SetLogLevel(*logLevel)
	if err != nil {
		log.Fatal().Msgf("Unknown log level: %s", *logLevel)
	}

	listenAddr := net.UDPAddr{Port: *port, IP: net.ParseIP("0.0.0.0")}

	// Create a tcp listener
	conn, err := net.ListenUDP("udp", &listenAddr)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error creating UDP listener on address %q", listenAddr)
	}
	log.Info().Msgf("Server listening on address %q", listenAddr)

	b := make([]byte, 2048)

	for {
		cc, remote, rderr := conn.ReadFromUDP(b)
		fmt.Println()
		fmt.Println(time.Now())
		fmt.Printf("Accepting a new packet\n")

		if rderr != nil {
			fmt.Printf("net.ReadFromUDP() error: %s\n", rderr)
		} else {
			fmt.Printf("Read %d bytes from socket\n", cc)
			fmt.Printf("Bytes: %q\n", string(b[:cc]))
		}

		fmt.Printf("Remote address: %v\n", remote)

		cc, wrerr := conn.WriteTo(b[0:cc], remote)
		if wrerr != nil {
			fmt.Printf("net.WriteTo() error: %s\n", wrerr)
		} else {
			fmt.Printf("Wrote %d bytes to socket\n", cc)
		}
	}
}
