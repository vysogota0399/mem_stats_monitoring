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

type Decryptor struct {
	privateKey *rsa.PrivateKey
}

func NewDecryptor(skpath string) (*Decryptor, error) {
	sk, err := os.ReadFile(skpath)
	if err != nil {
		return nil, fmt.Errorf("decryptor: failed to read private key: %w", err)
	}

	block, _ := pem.Decode(sk)
	if block == nil {
		return nil, fmt.Errorf("decryptor: failed to decode private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("decryptor: failed to parse private key: %w", err)
	}

	return &Decryptor{privateKey: privateKey.(*rsa.PrivateKey)}, nil
}

func (d *Decryptor) Decrypt(ciphertext string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("decryptor: failed to decode ciphertext: %w", err)
	}

	decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, d.privateKey, decoded, nil)
	if err != nil {
		return "", fmt.Errorf("decryptor: failed to decrypt ciphertext: %w", err)
	}

	return string(decrypted), nil
}
