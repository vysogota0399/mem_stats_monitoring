package keypair

import (
	"context"
	"crypto/x509/pkix"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/keypair/config"
	mocks "github.com/vysogota0399/mem_stats_monitoring/internal/mocks/keypair"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap/zapcore"
)

// Compile-time interface check
var _ IPemEncoder = (*PemEncoder)(nil)

func TestNewGenerator(t *testing.T) {
	// Create common logger once for all test cases
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: 0})
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	type args struct {
		cfg *config.Config
		lg  *logging.ZapLogger
	}
	tests := []struct {
		name string
		args args
		want *Generator
	}{
		{
			name: "successful generator creation",
			args: args{
				cfg: &config.Config{
					OutputFile: "test.pem",
					LogLevel:   zapcore.InfoLevel,
					Country:    "US",
					Province:   "CA",
					Locality:   "San Francisco",
					Org:        "Test Org",
					OrgUnit:    "Test Unit",
					CommonName: "test.example.com",
					Ttl:        time.Hour * 24 * 365,
				},
				lg: lg,
			},
			want: &Generator{
				cfg: &config.Config{
					OutputFile: "test.pem",
					LogLevel:   zapcore.InfoLevel,
					Country:    "US",
					Province:   "CA",
					Locality:   "San Francisco",
					Org:        "Test Org",
					OrgUnit:    "Test Unit",
					CommonName: "test.example.com",
					Ttl:        time.Hour * 24 * 365,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGenerator(tt.args.cfg, tt.args.lg)

			// Compare config fields
			if !reflect.DeepEqual(got.cfg, tt.want.cfg) {
				t.Errorf("NewGenerator() config = %v, want %v", got.cfg, tt.want.cfg)
			}

			// Verify logger is set
			if got.lg == nil {
				t.Error("NewGenerator() logger is nil")
			}
		})
	}
}

func TestGenerator_Call(t *testing.T) {
	// Create common logger once for all test cases
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: 0})
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	tests := []struct {
		name        string
		cfg         *config.Config
		mockEncoder *mocks.MockIPemEncoder
		wantErr     bool
		prepare     func(mockEncoder *mocks.MockIPemEncoder)
	}{
		{
			name: "successful keypair generation to stdout",
			cfg: &config.Config{
				LogLevel: zapcore.InfoLevel,
				Country:  "US",
				Ttl:      time.Hour * 24 * 365,
			},
			wantErr: false,
			prepare: func(mockEncoder *mocks.MockIPemEncoder) {
				mockEncoder.EXPECT().Encode(gomock.Any(), gomock.Any()).Return(nil).MaxTimes(2)
			},
		},
		{
			name: "encoding error",
			cfg: &config.Config{
				LogLevel: zapcore.InfoLevel,
				Country:  "US",
				Ttl:      time.Hour * 24 * 365,
			},
			prepare: func(mockEncoder *mocks.MockIPemEncoder) {
				mockEncoder.EXPECT().Encode(gomock.Any(), gomock.Any()).Return(errors.New("encoding error"))
				mockEncoder.EXPECT().Encode(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tt.mockEncoder = mocks.NewMockIPemEncoder(ctrl)
			tt.prepare(tt.mockEncoder)

			// Create generator with mock encoder
			g := &Generator{
				cfg: tt.cfg,
				lg:  lg,
				pe:  tt.mockEncoder,
			}

			// Run the test
			err := g.Call(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerator_newSubject(t *testing.T) {
	type fields struct {
		cfg *config.Config
		lg  *logging.ZapLogger
		pe  IPemEncoder
	}
	tests := []struct {
		name   string
		fields fields
		want   pkix.Name
	}{
		{
			name: "all fields populated",
			fields: fields{
				cfg: &config.Config{
					Country:    "US",
					Province:   "CA",
					Locality:   "San Francisco",
					Org:        "Test Org",
					OrgUnit:    "Test Unit",
					CommonName: "test.example.com",
				},
			},
			want: pkix.Name{
				Country:            []string{"US"},
				Province:           []string{"CA"},
				Locality:           []string{"San Francisco"},
				Organization:       []string{"Test Org"},
				OrganizationalUnit: []string{"Test Unit"},
				CommonName:         "test.example.com",
			},
		},
		{
			name: "only country field",
			fields: fields{
				cfg: &config.Config{
					Country: "US",
				},
			},
			want: pkix.Name{
				Country: []string{"US"},
			},
		},
		{
			name: "empty config",
			fields: fields{
				cfg: &config.Config{},
			},
			want: pkix.Name{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Generator{
				cfg: tt.fields.cfg,
				lg:  tt.fields.lg,
				pe:  tt.fields.pe,
			}
			if got := g.newSubject(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Generator.newSubject() = %v, want %v", got, tt.want)
			}
		})
	}
}
