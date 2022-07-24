package main

import (
	"errors"
	"log"

	"github.com/Tnze/go-mc/data/packetid"
	"github.com/Tnze/go-mc/net"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/google/uuid"
)

func handlePlaying(conn net.Conn, name string, uuid uuid.UUID, protocol int32) (err error) {
	serverConn, err := net.DialMC("localhost:25566")
	if err != nil {
		return errors.New("Error while connecting to backend server")
	}

	go serverPacketHandler(conn, serverConn, name)
	clientPacketHandler(conn, serverConn)
	return
}

// THESE ARE ALL SERVERBOUND PACKETS
func clientPacketHandler(conn net.Conn, serverConn *net.Conn) {
	for {
		var p pk.Packet
		err := conn.ReadPacket(&p)
		if err != nil {
			log.Printf("Serverbound: err: %v", err)
			serverConn.Close()
			return
		}
		switch p.ID {
		default:
			log.Printf("FROM CLIENT 0x%X", p.ID)
			serverConn.WritePacket(p)
		}
	}
}

// THESE ARE ALL CLIENTBOUND PACKETS + SERVER LOGIN LOGIC
func serverPacketHandler(conn net.Conn, serverConn *net.Conn, name string) {
	// 0 => login
	// 1 => play
	state := 0
	// Handshake
	err := serverConn.WritePacket(pk.Marshal(
		0x00,
		pk.VarInt(ProtocolVersion), // Protocol version
		pk.String("localhost"),     // Host
		pk.UnsignedShort(25566),    // Port
		pk.Byte(2),
	))
	if err != nil {
		log.Panicf("Error: %v", errors.New("handshake"))
	}
	// Login Start
	err = serverConn.WritePacket(pk.Marshal(
		packetid.LoginStart,
		pk.String(name),
		pk.Boolean(false),
	))
	if err != nil {
		log.Panicf("Error: %v", errors.New("login start"))
	}
	for {
		var p pk.Packet
		if err := serverConn.ReadPacket(&p); err != nil {
			log.Printf("Clientbound: Err: %v", err)
			return
		}
		if state == 0 {
			// Login packets
			switch p.ID {
			case 0x00:
				log.Panicln("Server disconnected client while login")
			case 0x01:
				log.Panicln("Backend server is in online mode!")
			case 0x02:
				log.Printf("Login success")
				state = 1
			case 0x03:
				log.Printf("Compression")
				var threshold pk.VarInt
				p.Scan(&threshold)
				log.Printf("Threshold: %x", threshold)
				serverConn.SetThreshold(int(threshold))
			}
		} else {
			// Play packets
			switch p.ID {
			default:
				log.Printf("FROM SERVER 0x%X", p.ID)
				conn.WritePacket(p)
			}
		}
	}
}

/*
func disconnectPlaying(conn net.Conn, message chat.Message) error {
	return conn.WritePacket(pk.Marshal(0x17,
		message))
}
*/
