package storages

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server/storages"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

func TestNewFileMetricsRestorer(t *testing.T) {
	type args struct {
		cfg        *config.Config
		srcBuilder *storages.MockSourceBuilder
		prepare    func(a *args)
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "when file storage path is empty",
			args: args{
				cfg: &config.Config{},
				prepare: func(a *args) {
				},
			},
			wantErr: true,
		},
		{
			name: "when file storage path found",
			args: args{
				cfg: &config.Config{FileStoragePath: "source"},
				prepare: func(a *args) {
					a.srcBuilder.EXPECT().Source(gomock.Any()).Return(&ReadCloserTarget{}, nil)
				},
			},
			wantErr: false,
		},
	}

	cntr := gomock.NewController(t)
	defer cntr.Finish()

	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, err)

			tt.args.srcBuilder = storages.NewMockSourceBuilder(cntr)
			tt.args.prepare(&tt.args)

			got, err := NewFileMetricsRestorer(tt.args.cfg, lg, nil, tt.args.srcBuilder) // TODO: implement me!!!!!!
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

type ReadCloserTarget struct {
	*bytes.Buffer
}

func (w *ReadCloserTarget) Close() error {
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
				source: &ReadCloserTarget{Buffer: &bytes.Buffer{}},
			},
			args: args{
				prepare: func(fields *fields) {},
			},
			wantErr: false,
		},
		{
			name: "when file is not valid",
			fields: fields{
				source: &ReadCloserTarget{Buffer: bytes.NewBufferString("invalid\n")},
			},
			args: args{
				prepare: func(fields *fields) {},
			},
			wantErr: true,
		},
		{
			name: "when save record to storage failed",
			fields: fields{
				source: &ReadCloserTarget{Buffer: bytes.NewBufferString("{\"type\":\"counter\",\"name\":\"test\",\"value\":1}\n")},
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
				source: &ReadCloserTarget{Buffer: bytes.NewBufferString("{\"type\":\"counter\",\"name\":\"test\",\"value\":1}\n")},
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
