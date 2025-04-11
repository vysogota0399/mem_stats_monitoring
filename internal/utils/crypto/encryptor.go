package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
)

type Encryptor struct {
	publicKey *rsa.PublicKey
}

func NewEncryptor(cert io.Reader) (*Encryptor, error) {
	certBytes, err := io.ReadAll(cert)
	if err != nil {
		return nil, fmt.Errorf("encryptor: failed to read public key: %w", err)
	}

	block, _ := pem.Decode(certBytes)
	if block == nil {
		return nil, fmt.Errorf("encryptor: failed to decode public key")
	}
	parsedCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("encryptor: failed to parse public key: %w", err)
	}

	publicKey, ok := parsedCert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("encryptor: failed to cast public key to rsa.PublicKey")
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
