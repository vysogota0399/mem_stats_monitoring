package storages

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server/storages"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

func TestNewFileMetricsRestorer(t *testing.T) {
	type args struct {
		cfg     *config.Config
		prepare func(cfg *config.Config) (*os.File, error)
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "when file storage path is empty",
			args: args{
				cfg: &config.Config{FileStoragePath: ""},
				prepare: func(cfg *config.Config) (*os.File, error) {
					return nil, nil
				},
			},
			wantErr: true,
		},
		{
			name: "when file found",
			args: args{
				cfg: &config.Config{},
				prepare: func(cfg *config.Config) (*os.File, error) {
					f, err := os.CreateTemp("", "dump.json")
					if err != nil {
						return nil, err
					}

					cfg.FileStoragePath = f.Name()
					return f, nil
				},
			},
			wantErr: false,
		},
	}

	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := tt.args.prepare(tt.args.cfg)
			assert.NoError(t, err)

			got, err := NewFileMetricsRestorer(tt.args.cfg, lg, nil)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}

			if f != nil {
				if err := f.Close(); err != nil {
					t.Error(err)
				}

				if err := os.Remove(f.Name()); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

type WriteCloserTarget struct {
	*bytes.Buffer
}

func (w *WriteCloserTarget) Close() error {
	return nil
}

func TestMetricsRestorer_Call(t *testing.T) {
	type fields struct {
		source io.ReadCloser
		strg   *storages.MockIStorageTarget
	}
	type args struct {
		prepare func(*fields)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "when empty file",
			fields: fields{
				source: &WriteCloserTarget{Buffer: &bytes.Buffer{}},
			},
			args: args{
				prepare: func(fields *fields) {},
			},
			wantErr: false,
		},
		{
			name: "when file is not valid",
			fields: fields{
				source: &WriteCloserTarget{Buffer: bytes.NewBufferString("invalid\n")},
			},
			args: args{
				prepare: func(fields *fields) {},
			},
			wantErr: true,
		},
		{
			name: "when save record to storage failed",
			fields: fields{
				source: &WriteCloserTarget{Buffer: bytes.NewBufferString("{\"type\":\"counter\",\"name\":\"test\",\"value\":1}\n")},
			},
			args: args{
				prepare: func(fields *fields) {
					fields.strg.EXPECT().Tx(gomock.Any(), gomock.Any()).Return(errors.New("tx error"))
				},
			},
			wantErr: true,
		},
		{
			name: "when success",
			fields: fields{
				source: &WriteCloserTarget{Buffer: bytes.NewBufferString("{\"type\":\"counter\",\"name\":\"test\",\"value\":1}\n")},
			},
			args: args{
				prepare: func(fields *fields) {
					fields.strg.EXPECT().Tx(gomock.Any(), gomock.Any()).Return(nil)
				},
			},
			wantErr: false,
		},
	}

	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	cntr := gomock.NewController(t)
	defer cntr.Finish()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.strg = storages.NewMockIStorageTarget(cntr)

			restorer := &MetricsRestorer{
				source: tt.fields.source,
				lg:     lg,
				cfg:    &config.Config{},
				strg:   tt.fields.strg,
			}

			tt.args.prepare(&tt.fields)
			err := restorer.Call(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
