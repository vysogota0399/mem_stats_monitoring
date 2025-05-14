package repositories

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mock "github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server/storages"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storages"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

func TestNewCounterRepository(t *testing.T) {
	type args struct {
		strg storages.Storage
		lg   *logging.ZapLogger
	}
	tests := []struct {
		name string
		args args
		want *CounterRepository
	}{
		{
			name: "when ok",
			args: args{
				strg: nil,
				lg:   nil,
			},
			want: &CounterRepository{
				storage: nil,
				lg:      nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCounterRepository(tt.args.strg, tt.args.lg)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCounterRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCounterRepository_Create(t *testing.T) {
	type fields struct {
		storage *mock.MockStorage
		actual  *models.Counter
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		prepare func(fields *fields)
		want    *models.Counter
	}{
		{
			name: "when get counter error",
			prepare: func(fields *fields) {
				fields.storage.EXPECT().GetCounter(gomock.Any(), gomock.Any()).Return(errors.New("get counter error"))
			},
			fields: fields{
				actual: &models.Counter{Name: "test", Value: 1},
			},
			wantErr: true,
		},
		{
			name: "when create or update counter error",
			prepare: func(fields *fields) {
				fields.storage.EXPECT().GetCounter(gomock.Any(), gomock.Any()).Return(nil)
				fields.storage.EXPECT().CreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("create or update counter error"))
			},
			fields: fields{
				actual: &models.Counter{Name: "test", Value: 1},
			},
			wantErr: true,
		},
		{
			name: "when ok",
			prepare: func(fields *fields) {
				fields.storage.EXPECT().GetCounter(gomock.Any(), gomock.Any())
				fields.storage.EXPECT().CreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, mType, mName string, val any) error {
					fields.actual.Value++
					return nil
				})
			},
			fields: fields{
				actual: &models.Counter{Name: "test", Value: 1},
			},
			wantErr: false,
			want:    &models.Counter{Name: "test", Value: 2},
		},
	}
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	cntr := gomock.NewController(t)
	defer cntr.Finish()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.storage = mock.NewMockStorage(cntr)
			tt.prepare(&tt.fields)

			rep := NewCounterRepository(tt.fields.storage, lg)
			err := rep.Create(context.Background(), tt.fields.actual)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.fields.actual.Value, tt.want.Value)
			}
		})
	}
}

func TestCounterRepository_FindByName(t *testing.T) {
	type fields struct {
		storage *mock.MockStorage
	}
	tests := []struct {
		name    string
		fields  fields
		prepare func(fields *fields)
		wantErr bool
	}{
		{
			name: "when get counter error",
			prepare: func(fields *fields) {
				fields.storage.EXPECT().GetCounter(gomock.Any(), gomock.Any()).Return(errors.New("get counter error"))
			},
			wantErr: true,
		},
		{
			name: "when found counter",
			prepare: func(fields *fields) {
				fields.storage.EXPECT().GetCounter(gomock.Any(), gomock.Any()).Return(nil)
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
			tt.fields.storage = mock.NewMockStorage(cntr)
			tt.prepare(&tt.fields)

			rep := NewCounterRepository(tt.fields.storage, lg)
			_, err := rep.FindByName(context.Background(), "test")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCounterRepository_All(t *testing.T) {
	type fields struct {
		storage *mock.MockStorage
	}
	tests := []struct {
		name    string
		fields  fields
		prepare func(fields *fields)
		wantErr bool
	}{
		{
			name: "when get counters error",
			prepare: func(fields *fields) {
				fields.storage.EXPECT().GetCounters(gomock.Any()).Return(nil, errors.New("get counters error"))
			},
			wantErr: true,
		},
		{
			name: "when found counters",
			prepare: func(fields *fields) {
				fields.storage.EXPECT().GetCounters(gomock.Any()).Return([]models.Counter{
					{Name: "test", Value: 1},
				}, nil)
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
			tt.fields.storage = mock.NewMockStorage(cntr)
			tt.prepare(&tt.fields)

			rep := NewCounterRepository(tt.fields.storage, lg)
			_, err := rep.All(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCounterRepository_SaveCollection(t *testing.T) {
	type fields struct {
		storage *mock.MockStorage
	}
	tests := []struct {
		name    string
		fields  fields
		prepare func(fields *fields)
		wantErr bool
	}{
		{
			name: "when search by name error",
			prepare: func(fields *fields) {
				fields.storage.EXPECT().GetCounter(gomock.Any(), gomock.Any()).Return(errors.New("get counter error"))
				fields.storage.EXPECT().Tx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
					return fn(ctx)
				})
			},
			wantErr: true,
		},
		{
			name: "when save collection error",
			prepare: func(fields *fields) {
				fields.storage.EXPECT().GetCounter(gomock.Any(), gomock.Any()).Return(nil)
				fields.storage.EXPECT().CreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("create or update counter error"))
				fields.storage.EXPECT().Tx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
					return fn(ctx)
				})
			},
			wantErr: true,
		},
		{
			name: "when ok",
			prepare: func(fields *fields) {
				fields.storage.EXPECT().Tx(gomock.Any(), gomock.Any()).Return(nil)
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
			tt.fields.storage = mock.NewMockStorage(cntr)
			tt.prepare(&tt.fields)

			rep := NewCounterRepository(tt.fields.storage, lg)
			err := rep.SaveCollection(context.Background(), []models.Counter{
				{Name: "test", Value: 1},
			})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCounterRepository_SearchByName(t *testing.T) {
	type fields struct {
		storage *mock.MockStorage
	}
	tests := []struct {
		name    string
		fields  fields
		prepare func(fields *fields)
		wantErr bool
	}{
		{
			name: "when search by name error",
			prepare: func(fields *fields) {
				fields.storage.EXPECT().GetCounter(gomock.Any(), gomock.Any()).Return(errors.New("get counter error"))
				fields.storage.EXPECT().Tx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
					return fn(ctx)
				})
			},
			wantErr: true,
		},
		{
			name: "when ok",
			prepare: func(fields *fields) {
				fields.storage.EXPECT().GetCounter(gomock.Any(), gomock.Any()).Return(nil)
				fields.storage.EXPECT().Tx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
					return fn(ctx)
				})
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
			tt.fields.storage = mock.NewMockStorage(cntr)
			tt.prepare(&tt.fields)

			rep := NewCounterRepository(tt.fields.storage, lg)
			_, err := rep.SearchByName(context.Background(), []string{"test"})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
