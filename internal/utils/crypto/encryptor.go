package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
)

type Encryptor struct {
	publicKey *rsa.PublicKey
}

func NewEncryptor(pkpath string) (*Encryptor, error) {
	pk, err := os.ReadFile(pkpath)
	if err != nil {
		return nil, fmt.Errorf("encryptor: failed to read public key: %w", err)
	}

	block, _ := pem.Decode(pk)
	if block == nil {
		return nil, fmt.Errorf("encryptor: failed to decode public key")
	}
	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("encryptor: failed to parse public key: %w", err)
	}

	return &Encryptor{publicKey: publicKey}, nil
}

func (e *Encryptor) Encrypt(message []byte) (string, error) {
	encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, e.publicKey, message, nil)
	if err != nil {
		return "", fmt.Errorf("encryptor: failed to encrypt message: %w", err)
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}
