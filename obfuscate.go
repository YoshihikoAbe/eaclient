package eaclient

import (
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rc4"
	"encoding/binary"
	"encoding/hex"
	"strings"
	"time"
)

type eamuseInfo [6]byte

func (info eamuseInfo) String() string {
	return "1-" + hex.EncodeToString(info[:4]) + "-" + hex.EncodeToString(info[4:])
}

func (info *eamuseInfo) fill() {
	binary.BigEndian.PutUint32(info[:], uint32(time.Now().Unix()))
	rand.Read(info[4:])
}

func (info *eamuseInfo) parse(s string) error {
	split := strings.Split(s, "-")
	if len(split) != 3 || split[0] != "1" {
		return clientError("malformed X-Eamuse-Info value")
	}
	_, err := hex.Decode(info[:], []byte(split[1]+split[2]))
	return err
}

func (info eamuseInfo) makeCipher() cipher.Stream {
	secret := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x69, 0xD7,
		0x46, 0x27, 0xD9, 0x85, 0xEE, 0x21, 0x87, 0x16,
		0x15, 0x70, 0xD0, 0x8D, 0x93, 0xB1, 0x24, 0x55,
		0x03, 0x5B, 0x6D, 0xF0, 0xD8, 0x20, 0x5D, 0xF5,
	}
	copy(secret, info[:])

	key := md5.Sum(secret)
	cipher, _ := rc4.NewCipher(key[:])
	return cipher
}
