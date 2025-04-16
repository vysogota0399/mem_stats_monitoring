package storages

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server/storages"
	"github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server/storages/dump"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

func TestNewPersistance(t *testing.T) {
	type args struct {
		dumper   *dump.MockDumper
		restorer *storages.MockRestorer
	}
	tests := []struct {
		name         string
		args         args
		want         *Persistance
		prepare      func(args *args)
		wantStartErr bool
		wantStopErr  bool
	}{
		{
			name: "when start and restore failed",
			args: args{},
			prepare: func(args *args) {
				args.restorer.EXPECT().Call(gomock.Any()).Return(errors.New("restore failed"))
			},
			wantStartErr: true,
		},
		{
			name: "when start and restore succeded and stop succeded",
			args: args{},
			prepare: func(args *args) {
				args.restorer.EXPECT().Call(gomock.Any()).Return(nil)
				args.dumper.EXPECT().Start(gomock.Any())
				args.dumper.EXPECT().Stop(gomock.Any())
			},
		},
	}

	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	cntr := gomock.NewController(t)
	defer cntr.Finish()

	f, err := os.CreateTemp("", "test.json")
	assert.NoError(t, err)
	defer func() {
		if err := f.Close(); err != nil {
			lg.ErrorCtx(context.Background(), "close error", zap.Error(err))
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var (
				l fx.Lifecycle
				s fx.Shutdowner
			)
			app := fxtest.New(
				t,
				fx.Populate(&l, &s),
			)
			tt.args.dumper = dump.NewMockDumper(cntr)
			tt.args.restorer = storages.NewMockRestorer(cntr)
			tt.prepare(&tt.args)

			got, err := NewPersistance(l, &config.Config{FileStoragePath: f.Name()}, tt.args.dumper, lg)
			assert.NotNil(t, got)
			assert.NoError(t, err)

			got.restorer = tt.args.restorer

			err = app.Start(context.Background())

			if tt.wantStartErr {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}

			stopErr := app.Stop(context.Background())
			if tt.wantStopErr {
				assert.Error(t, stopErr)
				return
			} else {
				assert.NoError(t, stopErr)
			}
		})
	}
}

func TestPersistance_CreateOrUpdate(t *testing.T) {
	type fields struct {
		Memory *Memory
		dumper *dump.MockDumper
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		prepare func(fields *fields)
	}{
		{
			name: "when ok",
			prepare: func(fields *fields) {
				fields.dumper.EXPECT().Dump(gomock.Any(), gomock.Any())
			},
		},
	}

	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	cntr := gomock.NewController(t)
	defer cntr.Finish()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.dumper = dump.NewMockDumper(cntr)
			tt.prepare(&tt.fields)

			p := &Persistance{
				Memory:   NewMemory(lg),
				lg:       lg,
				dumper:   tt.fields.dumper,
				restorer: nil,
			}

			err := p.CreateOrUpdate(context.Background(), models.CounterType, "mName", 1)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
