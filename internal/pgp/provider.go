package pgp

import "golang.org/x/crypto/openpgp/packet"

type PGPProvider interface {
	GetPublicKey() *packet.PublicKey
}
