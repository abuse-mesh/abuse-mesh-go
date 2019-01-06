package pgp

import (
	"bytes"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

func PGPEntityFromBytes(packetBytes []byte) (*openpgp.Entity, error) {
	//Create a buffer from the pgp packet bytes
	packetBuffer := bytes.NewBuffer(packetBytes)

	//wrap the buffer with a reader which decodes the pgp packets
	packetReader := packet.NewReader(packetBuffer)

	return openpgp.ReadEntity(packetReader)
}
