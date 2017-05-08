package keyman

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
)

// Key contains the full key together with its fingerprint
type Key struct {
	bytes       []byte
	fingerprint [32]byte // The SHA-256 fingerprint of this key.
}

// NewKey generates a new key from a set of bytes. If the number of bytes does
// not equal 16, 24 or 32, an error is returned
func NewKey(data []byte) (*Key, error) {
	switch len(data) {
	case 16, 24, 32:
		break
	default:
		return nil, aes.KeySizeError(len(data))
	}

	return &Key{
		bytes:       data,
		fingerprint: sha256.Sum256(data),
	}, nil
}

// GenerateKey creates a new key of the specified size. If the size does not
// equal 16, 24 or 32, an error is returned
func GenerateKey(size int) (*Key, error) {
	switch size {
	case 16, 24, 32:
		break
	default:
		return nil, aes.KeySizeError(size)
	}
	data := make([]byte, size)
	rand.Read(data)
	return NewKey(data)
}

// GetBytes returns the bytes of the key
func (key Key) GetBytes() []byte {
	return key.bytes
}

// GetFingerprint returns the fingerprint of the key
func (key Key) GetFingerprint() [32]byte {
	return key.fingerprint
}
