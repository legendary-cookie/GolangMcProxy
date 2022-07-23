package main

import (
	"fmt"

	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data/packetid"
	"github.com/Tnze/go-mc/net"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/Tnze/go-mc/offline"
	"github.com/Tnze/go-mc/server/auth"
	"github.com/google/uuid"
)

type PlayerInfo struct {
	Name string
	UUID uuid.UUID
}

func Login(conn net.Conn, conf *Config) (name string, uuid uuid.UUID, profilePubKey *auth.PublicKey, properties []auth.Property, err error) {
	//login start
	var p pk.Packet
	err = conn.ReadPacket(&p)
	if err != nil {
		return
	}
	if p.ID != packetid.LoginStart {
		err = wrongPacketErr{expect: packetid.LoginStart, get: p.ID}
		return
	}

	var hasPubKey pk.Boolean
	var pubKey auth.PublicKey
	err = p.Scan(
		(*pk.String)(&name),
		&hasPubKey,
		pk.Opt{
			Has:   &hasPubKey,
			Field: &pubKey,
		},
	) //decode username as pk.String
	if err != nil {
		return
	}

	if hasPubKey {
		if !pubKey.Verify() {
			err = LoginFailErr{reason: chat.TranslateMsg("multiplayer.disconnect.invalid_public_key_signature")}
			return
		}
		profilePubKey = &pubKey
	} else if conf.EnforceSecureProfile {
		err = LoginFailErr{reason: chat.TranslateMsg("multiplayer.disconnect.missing_public_key")}
		return
	}

	//auth
	if conf.OnlineMode {
		var resp *auth.Resp
		resp, err = auth.Encrypt(&conn, name, profilePubKey.PubKey)
		if err != nil {
			return
		}
		name = resp.Name
		uuid = resp.ID
		properties = resp.Properties
	} else {
		// offline-mode UUID
		uuid = offline.NameToUUID(name)
	}

	//set compression
	if conf.Threshold >= 0 {
		err = conn.WritePacket(pk.Marshal(
			packetid.LoginCompression,
			pk.VarInt(conf.Threshold),
		))
		if err != nil {
			return
		}
		conn.SetThreshold(conf.Threshold)
	}

	// send login success
	err = conn.WritePacket(pk.Marshal(
		packetid.LoginSuccess,
		pk.UUID(uuid),
		pk.String(name),
		pk.Array(properties),
	))

	return
}

type wrongPacketErr struct {
	expect, get int32
}

func (w wrongPacketErr) Error() string {
	return fmt.Sprintf("wrong packet id: expect %#02X, get %#02X", w.expect, w.get)
}

type LoginFailErr struct {
	reason chat.Message
}

func (l LoginFailErr) Error() string {
	return "login error: " + l.reason.ClearString()
}
