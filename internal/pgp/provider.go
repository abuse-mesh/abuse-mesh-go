package pgp

import (
	"golang.org/x/crypto/openpgp"
)

type PGPProvider interface {
	GetEntity() *openpgp.Entity
}
