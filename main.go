package main

import (
	"log"

	"github.com/Tnze/go-mc/net"
)

const ProtocolVersion = 759
const MaxPlayer = 50

type Config struct {
	OnlineMode           bool
	EnforceSecureProfile bool
	Threshold            int
}

func main() {
	conf := &Config{
		OnlineMode:           false,
		EnforceSecureProfile: false,
		Threshold:            100,
	}
	l, err := net.ListenMC(":25565")
	if err != nil {
		log.Fatalf("Listen error: %v", err)
	}
	log.Printf("Proxy listening for incoming connections")
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalf("Accept error: %v", err)
		}
		go acceptConn(conn, conf)
	}
}

func acceptConn(conn net.Conn, conf *Config) {
	defer conn.Close()
	protocol, intention, _, err := Handshake(conn)
	if err != nil {
		log.Printf("Handshake error: %v", err)
		return
	}
	switch intention {
	default: //unknown error
		log.Printf("Unknown handshake intention: %v", intention)
	case 1: //for status
		StatusHandler(conn)
	case 2: //for login
		name, uuid, _, _, err := Login(conn, conf)
		if err != nil {
			log.Printf("Error while logging in: %v", err)
			return
		}
		log.Printf("User %s connected with UUID %s", name, uuid.String())
		// proxy to actual servers
		handlePlaying(conn, name, uuid, protocol)
	}
}
