package crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type hasher struct {
	response []byte
}

func newHasher(response string) *hasher {
	return &hasher{[]byte(response)}
}

func (h *hasher) Write(b []byte) (int, error) {
	return len(b), nil
}

func (h *hasher) Sum(b []byte) []byte {
	return h.response
}
func (h *hasher) Reset() {}

func (h *hasher) Size() int {
	return len(h.response)
}

func (h *hasher) BlockSize() int {
	return 0
}

func TestCms_Sign(t *testing.T) {
	type args struct {
		msg io.Reader
	}
	tests := []struct {
		name    string
		cms     *Cms
		args    args
		wantErr bool
		want    []byte
	}{
		{
			name:    "returns signature",
			cms:     NewCms(newHasher("response")),
			args:    args{bytes.NewBuffer([]byte("secret message"))},
			wantErr: false,
			want:    []byte("response"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.cms
			sign, err := c.Sign(tt.args.msg)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, sign)
		})
	}
}

func TestCms_Verify(t *testing.T) {
	const validKey = "valid_key"
	const invalidKey = "invalid_key"

	type args struct {
		secret       []byte
		data         []byte
		expectedSign func(io.Reader) ([]byte, error)
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "when invalid signature",
			args: args{
				secret: []byte(invalidKey),
				data:   []byte("strint to be signed"),
				expectedSign: func(b io.Reader) ([]byte, error) {
					return NewCms(hmac.New(sha256.New, []byte(validKey))).Sign(b)
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "when valid signature",
			args: args{
				secret: []byte(validKey),
				data:   []byte("string to be signed"),
				expectedSign: func(b io.Reader) ([]byte, error) {
					return NewCms(hmac.New(sha256.New, []byte(validKey))).Sign(b)
				},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cms := NewCms(hmac.New(sha256.New, tt.args.secret))
			expectedSign, err := tt.args.expectedSign(bytes.NewBuffer(tt.args.data))
			assert.NoError(t, err)
			valid, err := cms.Verify(bytes.NewBuffer(tt.args.data), expectedSign)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, valid)
		})
	}
}
