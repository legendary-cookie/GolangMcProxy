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
	err = serverConn.WritePacket(pk.Marshal(
		0x00,
		pk.VarInt(ProtocolVersion), // Protocol version
		pk.String("localhost"),     // Host
		pk.UnsignedShort(25566),    // Port
		pk.Byte(2),
	))
	if err != nil {
		return errors.New("handshake")
	}
	go serverPacketHandler(conn, serverConn)
	// Login Start
	err = serverConn.WritePacket(pk.Marshal(
		packetid.LoginStart,
		pk.String(name),
		pk.Boolean(false),
	))
	if err != nil {
		return errors.New("login start")
	}

	// Proxy stuff from client -> server
	for {
		var p pk.Packet
		err := conn.ReadPacket(&p)
		if err != nil {
			log.Printf("Serverbound: err: %v", err)
		}
		log.Printf("Serverbound: 0x%X", p.ID)
	}
}

func serverPacketHandler(conn net.Conn, serverConn *net.Conn) {
	for {
		var p pk.Packet
		if err := serverConn.ReadPacket(&p); err != nil {
			log.Printf("Err: %v", err)
			return
		}

	}
}
