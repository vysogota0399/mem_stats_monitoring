package storages

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server/storages"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestNewStorage(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	cntr := gomock.NewController(t)
	defer cntr.Finish()

	type args struct {
		connectionOpener *storages.MockConnectionOpener
		sourceBuilder    *storages.MockSourceBuilder
	}

	tests := []struct {
		name    string
		cfg     *config.Config
		want    Storage
		args    args
		prepare func(args *args)
		wantErr bool
	}{
		{
			name: "create memory storage",
			cfg: &config.Config{
				DatabaseDSN:     "",
				FileStoragePath: "",
			},
			args:    args{},
			prepare: func(args *args) {},
			want:    &Memory{},
			wantErr: false,
		},
		{
			name: "create postgres storage",
			cfg: &config.Config{
				DatabaseDSN:     "postgres://user:pass@localhost:5432/db",
				FileStoragePath: "",
			},
			want: &PG{},
			args: args{},
			prepare: func(args *args) {
				args.connectionOpener.EXPECT().OpenDB(gomock.Any()).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "create persistence storage",
			cfg: &config.Config{
				DatabaseDSN:     "",
				FileStoragePath: "/tmp/data",
			},
			want:    &Persistance{},
			wantErr: false,
			args:    args{},
			prepare: func(args *args) {
				args.sourceBuilder.EXPECT().Source(gomock.Any())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				l fx.Lifecycle
				s fx.Shutdowner
			)
			fxtest.New(
				t,
				fx.Populate(&l, &s),
			)

			tt.args.connectionOpener = storages.NewMockConnectionOpener(cntr)
			tt.args.sourceBuilder = storages.NewMockSourceBuilder(cntr)
			tt.prepare(&tt.args)

			got, err := NewStorage(l, nil, lg, tt.cfg, nil, tt.args.connectionOpener, tt.args.sourceBuilder)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.IsType(t, tt.want, got)
			}
		})
	}
}
