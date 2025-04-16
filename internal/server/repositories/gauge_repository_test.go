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

func TestNewGaugeRepository(t *testing.T) {
	type args struct {
		strg storages.Storage
		lg   *logging.ZapLogger
	}
	tests := []struct {
		name string
		args args
		want *GaugeRepository
	}{
		{
			name: "when ok",
			args: args{
				strg: nil,
				lg:   nil,
			},
			want: &GaugeRepository{
				storage: nil,
				lg:      nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGaugeRepository(tt.args.strg, tt.args.lg)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGaugeRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGaugeRepository_Create(t *testing.T) {
	type fields struct {
		storage *mock.MockStorage
		actual  *models.Gauge
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		prepare func(fields *fields)
	}{
		{
			name: "when create or update gauge error",
			prepare: func(fields *fields) {
				fields.storage.EXPECT().CreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("create or update gauge error"))
			},
			fields: fields{
				actual: &models.Gauge{Name: "test", Value: 1.0},
			},
			wantErr: true,
		},
		{
			name: "when ok",
			prepare: func(fields *fields) {
				fields.storage.EXPECT().CreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			fields: fields{
				actual: &models.Gauge{Name: "test", Value: 1.0},
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

			rep := NewGaugeRepository(tt.fields.storage, lg)
			err := rep.Create(context.Background(), tt.fields.actual)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGaugeRepository_FindByName(t *testing.T) {
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
			name: "when get gauge error",
			prepare: func(fields *fields) {
				fields.storage.EXPECT().GetGauge(gomock.Any(), gomock.Any()).Return(errors.New("get gauge error"))
			},
			wantErr: true,
		},
		{
			name: "when found gauge",
			prepare: func(fields *fields) {
				fields.storage.EXPECT().GetGauge(gomock.Any(), gomock.Any()).Return(nil)
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

			rep := NewGaugeRepository(tt.fields.storage, lg)
			_, err := rep.FindByName(context.Background(), "test")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGaugeRepository_All(t *testing.T) {
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
			name: "when get gauges error",
			prepare: func(fields *fields) {
				fields.storage.EXPECT().GetGauges(gomock.Any()).Return(nil, errors.New("get gauges error"))
			},
			wantErr: true,
		},
		{
			name: "when found gauges",
			prepare: func(fields *fields) {
				fields.storage.EXPECT().GetGauges(gomock.Any()).Return([]models.Gauge{
					{Name: "test", Value: 1.0},
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

			rep := NewGaugeRepository(tt.fields.storage, lg)
			_, err := rep.All(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGaugeRepository_SaveCollection(t *testing.T) {
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
			name: "when save collection error",
			prepare: func(fields *fields) {
				fields.storage.EXPECT().Tx(gomock.Any(), gomock.Any()).Return(errors.New("save collection error"))
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

			rep := NewGaugeRepository(tt.fields.storage, lg)
			err := rep.SaveCollection(context.Background(), []models.Gauge{
				{Name: "test", Value: 1.0},
			})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
