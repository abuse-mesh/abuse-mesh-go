package pgp

import (
	"encoding/binary"
	"encoding/hex"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/openpgp"
)

//FileGPGProvider provides pgp via files on the filesystem
type FilePGPProvider struct {
	entityList      openpgp.EntityList
	keyringLocation string
	keyID           uint64
	keyInUse        openpgp.Key
	passphrase      []byte
}

func (provider *FilePGPProvider) GetEntity() *openpgp.Entity {
	return provider.keyInUse.Entity
}

func (provider *FilePGPProvider) init() error {
	keyring, err := os.Open(provider.keyringLocation)
	defer keyring.Close()
	if err != nil {
		return errors.Wrap(err, "Error while opening pgp keyring")
	}

	provider.entityList, err = openpgp.ReadKeyRing(keyring)
	if err != nil {
		return errors.Wrap(err, "Error while reading pgp keyring")
	}

	keys := provider.entityList.KeysById(provider.keyID)
	if len(keys) != 1 {
		return errors.Errorf("Found '%d' keys in the keyring matching id '%#x'", len(keys), provider.keyID)
	}

	key := keys[0]
	log.Infof("Keyring opened successfully, using PGP key with fingerprint '%#x'", key.PublicKey.Fingerprint)

	if key.PrivateKey == nil {
		return errors.New("Selected PGP key doesn't have a private key in the keyring")
	}

	if key.PrivateKey.Encrypted {
		log.Info("PGP private key is encrypted with a passphrase, attempting decryption")
		err := key.PrivateKey.Decrypt(provider.passphrase)
		if err != nil {
			log.Error("PGP decryption failed, passphrase invalid")
			return errors.Wrap(err, "Error while decrypting PGP private key")
		}
	}

	log.Info("PGP keypair loaded, ready to sign messages")

	provider.keyInUse = key

	return nil
}

func NewFilePGPProvider(keyringLocation, keyidString string, passphrase []byte) (*FilePGPProvider, error) {
	keyData, err := hex.DecodeString(keyidString)
	if err != nil {
		return nil, errors.Wrap(err, "Error while decoding PGP key id")
	}

	//Crate a key id
	keyid := binary.BigEndian.Uint64(keyData)

	provider := &FilePGPProvider{
		keyringLocation: keyringLocation,
		keyID:           keyid,
		passphrase:      passphrase,
	}

	err = provider.init()
	if err != nil {
		return nil, err
	}

	return provider, nil
}
