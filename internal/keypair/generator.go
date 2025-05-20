// Package keypair provides functionality for generating RSA key pairs and X.509 certificates.
// It supports configurable certificate attributes and concurrent generation of private keys
// and certificates. The package uses the crypto/rsa package for key generation and
// crypto/x509 for certificate creation.
//
// The main component is the Generator type which handles the key pair generation process.
// It can be configured with various certificate attributes like country, organization,
// common name, etc. The generation process is context-aware and can be cancelled.
//
// Example usage:
//
//	cfg := &config.Config{...}
//	lg := logging.NewZapLogger()
//	generator := NewGenerator(cfg, lg)
//	err := generator.Call(ctx)
package keypair

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"os"
	"time"

	"github.com/vysogota0399/mem_stats_monitoring/internal/keypair/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Generator struct {
	cfg *config.Config
	lg  *logging.ZapLogger
	pe  IPemEncoder
}

func NewGenerator(cfg *config.Config, lg *logging.ZapLogger) *Generator {
	return &Generator{cfg: cfg, lg: lg, pe: &PemEncoder{}}
}

type IPemEncoder interface {
	Encode(out io.Writer, block *pem.Block) error
}

type PemEncoder struct{}

func (p *PemEncoder) Encode(out io.Writer, block *pem.Block) error {
	return pem.Encode(out, block)
}

func (g *Generator) Call(ctx context.Context) error {
	g.lg.DebugCtx(
		ctx,
		"generating keypair",
		zap.Any("config", g.cfg),
	)

	pk, err := g.generateKey()
	if err != nil {
		return fmt.Errorf("generator: failed to generate key: %w", err)
	}

	cert := g.generateCert()
	errgroup, ctx := errgroup.WithContext(ctx)

	go g.export(
		ctx,
		errgroup,
		g.genCertPem(ctx, errgroup, &cert, pk),
		g.genKeyPem(ctx, errgroup, pk),
	)

	if err := errgroup.Wait(); err != nil {
		return fmt.Errorf("generator: failed to wait for errgroup: %w", err)
	}

	return nil
}

func (g *Generator) export(ctx context.Context, errgroup *errgroup.Group, buffers ...chan *bytes.Buffer) {
	for _, bCH := range buffers {
		errgroup.Go(func() error {
			b := <-bCH

			if g.cfg.OutputFile == "" {
				select {
				case <-ctx.Done():
					return fmt.Errorf("generator: context done")
				default:
					if _, err := os.Stdout.Write(b.Bytes()); err != nil {
						return fmt.Errorf("generator: failed to write to stdout: %w", err)
					}
				}

				return nil
			}

			f, err := os.Create(g.cfg.OutputFile)
			if err != nil {
				return fmt.Errorf("generator: failed to create output file: %w", err)
			}
			defer func() {
				if err := f.Close(); err != nil {
					g.lg.ErrorCtx(ctx, "failed to close file", zap.Error(err))
				}
			}()

			select {
			case <-ctx.Done():
				return fmt.Errorf("generator: context done")
			default:
				if _, err := f.Write(b.Bytes()); err != nil {
					return fmt.Errorf("generator: failed to write to output file: %w", err)
				}
			}

			return nil
		})
	}
}

func (g *Generator) genCertPem(ctx context.Context, errgroup *errgroup.Group, cert *x509.Certificate, pk *rsa.PrivateKey) chan *bytes.Buffer {
	result := make(chan *bytes.Buffer)

	errgroup.Go(func() error {
		defer close(result)

		certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &pk.PublicKey, pk)
		if err != nil {
			return fmt.Errorf("generator: failed to create certificate: %w", err)
		}

		var certPEM bytes.Buffer
		if err := g.pe.Encode(&certPEM, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
			return fmt.Errorf("generator: failed to encode certificate: %w", err)
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("generator: context done")
		case result <- &certPEM:
			return nil
		}
	})

	return result
}

func (g *Generator) genKeyPem(ctx context.Context, errgroup *errgroup.Group, pk *rsa.PrivateKey) chan *bytes.Buffer {
	result := make(chan *bytes.Buffer)

	errgroup.Go(func() error {
		defer close(result)

		keyBytes, err := x509.MarshalPKCS8PrivateKey(pk)
		if err != nil {
			return fmt.Errorf("generator: failed to marshal private key: %w", err)
		}

		var keyPEM bytes.Buffer
		if err := g.pe.Encode(&keyPEM, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes}); err != nil {
			return fmt.Errorf("generator: failed to encode key: %w", err)
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("generator: context done")
		case result <- &keyPEM:
			return nil
		}
	})

	return result
}

func (g *Generator) generateCert() x509.Certificate {
	return x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject:      g.newSubject(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(g.cfg.Ttl),
	}
}

func (g *Generator) generateKey() (*rsa.PrivateKey, error) {
	pk, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("generator: failed to generate key: %w", err)
	}
	return pk, nil
}

func (g *Generator) newSubject() pkix.Name {
	s := pkix.Name{}
	if g.cfg.Country != "" {
		s.Country = []string{g.cfg.Country}
	}
	if g.cfg.Province != "" {
		s.Province = []string{g.cfg.Province}
	}
	if g.cfg.Locality != "" {
		s.Locality = []string{g.cfg.Locality}
	}
	if g.cfg.Org != "" {
		s.Organization = []string{g.cfg.Org}
	}
	if g.cfg.OrgUnit != "" {
		s.OrganizationalUnit = []string{g.cfg.OrgUnit}
	}
	if g.cfg.CommonName != "" {
		s.CommonName = g.cfg.CommonName
	}
	return s
}
