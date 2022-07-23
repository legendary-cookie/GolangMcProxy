package main

import (
	"encoding/json"
	"log"

	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/net"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/google/uuid"
)

type player struct {
	Name string    `json:"name"`
	ID   uuid.UUID `json:"id"`
}

func listResp() string {
	var list struct {
		Version struct {
			Name     string `json:"name"`
			Protocol int    `json:"protocol"`
		} `json:"version"`
		Players struct {
			Max    int      `json:"max"`
			Online int      `json:"online"`
			Sample []player `json:"sample"`
		} `json:"players"`
		Description chat.Message `json:"description"`
		FavIcon     string       `json:"favicon,omitempty"`
	}

	list.Version.Name = "1.19"
	list.Version.Protocol = ProtocolVersion
	list.Players.Max = MaxPlayer
	list.Players.Online = 0
	list.Players.Sample = []player{} // must init. can't be nil
	list.Description = chat.Message{Text: "Custom proxy powered by go-mc", Color: "blue"}

	data, err := json.Marshal(list)
	if err != nil {
		log.Panic("Marshal JSON for status checking fail")
	}
	return string(data)
}

func StatusHandler(conn net.Conn) {
	var p pk.Packet
	for i := 0; i < 2; i++ { // ping or list. Only accept twice
		err := conn.ReadPacket(&p)
		if err != nil {
			return
		}
		switch p.ID {
		case 0x00: //List
			err = conn.WritePacket(pk.Marshal(0x00, pk.String(listResp())))
		case 0x01: //Ping
			err = conn.WritePacket(p)
		}
		if err != nil {
			return
		}
	}
}
