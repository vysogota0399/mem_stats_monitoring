package crypto

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncrypt(t *testing.T) {
	type args struct {
		cert    io.Reader
		message []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "successful encryption",
			args: args{
				message: []byte(`Hello, World!`),
				cert: bytes.NewReader([]byte(`-----BEGIN CERTIFICATE-----
MIIEfDCCAmSgAwIBAgIEZ/kQFzANBgkqhkiG9w0BAQsFADAAMB4XDTI1MDQxMTEy
NTAzMVoXDTI2MDQxMTEyNTAzMVowADCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCC
AgoCggIBALCMsAGE+KVp/Cb8q6+adz1zmfIOAYG0IYuLs8Mp+/5867F9cDNzAsi5
dkzVpnXPuqfRaJm9u3dAoJkQPBn8a+ggC0TGlqf7JjG/jdnPEzHrme5UML9Y2a1o
BG+PpqfLW8BLBVUfT32WAJilvIgQa7lQl1VZ59kdcsFkJI8YymSqcI22GwL0SRZv
sXWW2rWjixP23Pu35s2upRkqKw/OIB1s07uvTBNz/nuiCcmmcz76MOQYcDpz7UvQ
5vtIXaSLJd6zN/vZXF2KGemXnCLmGL8u3tJA+BWidsPqIcuWApWSs5G79gM2tQei
KCMcqbSbJ8XGLmQs/9rv9lvQkBCDbZfLBegQ7iWJxpFhxje8ysJ2A56/dqR2wKS2
Q9jK0xm/ARKR+h3zDZRhcXnQszLWzCstnf9OIaSTd12Oe1D3bbRC/ge5PW/5WQDH
ijxXpBU0lXXyJXxb+rFQcC7HfqKl/QDamd9DYkRZMA5U70p5h6h4/WFNvIhngCvl
eJ7M2O7XyCyhiCYliE5rpVm7KkD/QSCw2rTFESCovXLvkECbgS+qrg8QjIuNvYJs
J5xhc0bCdew9BP46A+oZKm7ByZNmJkjGKMIc3oFBNVUPCzS0PQ7R7yM9Vcly2yar
0hqbKT92Fucg/QFM1kAt5xjK+4TSd2gcrz8ShcqLSc/YwvWWLQxxAgMBAAEwDQYJ
KoZIhvcNAQELBQADggIBAEgI8WNcg2hA4YucvErJi+LRVZpkD7h4LaHpvtV+4ngg
fATvP4u8I3lxz4k1xhTaD3VEVwhDP4cainNeMPbzzYRv4WJ17Fews/hizzKJOP/d
n8/RNfGcDqxj2+SA4kDWaBy5r2czQyTeH05GO9GaTtZiCd34h5VtWh2ILvRcIrDJ
3H+qqRyJOA04urIj191y+01WRPq/uzsxbJFj+iO0pM8o4ROFLHwG3E0nz2GSBcC8
bfA1XxMejXJujS5Szl1NogAH47bnSIkqBKiQhytMyQPqFEb/QkdrB314qUrR0rNZ
37w4U2CCa5C/N8FX3vlhEPNXTmotEhkO05MZfPYkv/q8gPJVr1fM58NLyouFgk2y
Ji/i1QIN8DQliysJLqK2h2G9gbT1vrzNtTuIRs2mnAJQD2PEYYMGpEbAOGhu9mkz
9sTS07kZdb+DjqNvEemknacSIAWqvpHwzP6PDROjOVxXf8+grCBiFLTUV6Xkf8V8
kUKrzPIkDfRQYxHh0kna2Up4vPawLP3ZStAVzjskUUHmjZz0SX8lORbv1fCmSJct
Wc774xRQem2kUm1sVcuy3LV2xpxSNSksrcsM10Rs4RdOiuysnHq/bLySX6tVe4ib
IOLlRE1QeXDJuz9kAbWJSk0+ZkpGornCXCqI8TLOJ1XpcQG93A8DIM6n+jx8v/n/
-----END CERTIFICATE-----
				`)),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encryptor, err := NewEncryptor(tt.args.cert)
			assert.NoError(t, err)

			message, err := encryptor.Encrypt(tt.args.message)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			fmt.Println(message)
			assert.NotEqual(t, "", message)
		})
	}
}
