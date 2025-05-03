package crypto

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
)

func TestDecryptor_Decrypt(t *testing.T) {
	type fields struct {
		privateKey io.Reader
	}
	type args struct {
		ciphertext string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "successful decryption",
			want:    "Hello, World!",
			wantErr: false,
			fields: fields{
				privateKey: bytes.NewReader([]byte(`-----BEGIN RSA PRIVATE KEY-----
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
yxKZlQ/bpyBAbIhe9BQ1j6cjfQo+
-----END RSA PRIVATE KEY-----
`)),
			},
			args: args{
				ciphertext: "AteuC1yLSWz7x5gI+wddlAZT7UDCVxILSeCstlP77vSKLG/nXA7302xj5ZqRK2q5wpeXQEAvJL2GVgQmNXcGfdo+WvjgJt+t2oyVG7LwTf95h4B5mRlnIHpXK5m6oGLSl36oAS9O5cGZTRsK101zBBEo+tTPjC3QH99VrKa6qympb5udSglvxyy4x6S7dI3TMAM413qMZDPfS9YCLKqEF1cMTw3ItGIr5ZtevFdAOqdXnvkoqAnE3tLQgaT5D5R+Xd9qmi6+jDlK/1x6hdLyCqdbxQiaienDebSMvjeEIdBKX8i+tTgcQsIgGwIvujRHRS7KJQg/gcb+DGFRb6FDTkgNdHER726x6zNU2wGaY37h/wRyyBnGz5KBkywycxy6rxtA3ffVsee+9dy49DdfM5m3V6C6uMyozC4lJdTUCedct4m9/+f8SEzNk9UFWYcOm/sQk6nucQbZ2vlNu5w/++6/G00ReFLvy8AlgO06AQIyM6no7MyFPF+apwT1tDEOAmLEDevLjPXD89IjcLq90r/rggX1bpqIWid4/rbA5Iyc5Wki9V2sHyMq+0ghxxNMR11qFf2UJgn44BVYuahT17QJuOFugAtmBJVeG2S33+rTtkppDaX3QQ2zgoqaWd0M5mBihNxv47PvZCcfNnWZ0vd/daNNreDaIQwzHY/zJ6M=",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decryptor := NewDecryptor(&config.Config{PrivateKey: tt.fields.privateKey})

			got, err := decryptor.Decrypt(tt.args.ciphertext)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
