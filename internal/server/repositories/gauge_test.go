package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	mocks "github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

func TestGauge_Last(t *testing.T) {
	type args struct {
		storage storage.Storage
		mName   string
	}
	tests := []struct {
		name    string
		args    args
		want    *models.Gauge
		wantErr error
	}{
		{
			name: "when storage has record then returns record",
			args: args{
				mName: "testSetGet241",
				storage: storage.NewMemStorageWithData(
					map[string]map[string][]string{
						"gauge": {
							"testSetGet241": []string{`{"value": 120400.951, "name": "testSetGet241"}`},
						},
					},
				),
			},
			want:    &models.Gauge{Value: 120400.951, Name: "testSetGet241"},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			g := NewGauge(tt.args.storage, lg)
			got, err := g.Last(context.Background(), tt.args.mName)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestGauge_All(t *testing.T) {
	type fields struct {
		storage storage.Storage
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string][]models.Gauge
	}{
		{
			name: "return all records",
			fields: fields{
				storage: storage.NewMemStorageWithData(
					map[string]map[string][]string{
						"gauge": {
							"testSetGet241": []string{`{"value": 120400.951, "name": "testSetGet241"}`},
						},
					},
				),
			},
			want: map[string][]models.Gauge{
				"testSetGet241": {{Value: 120400.951, Name: "testSetGet241"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			g := Gauge{
				storage: tt.fields.storage,
				lg:      lg,
			}
			got := g.All()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGauge_SaveCollection(t *testing.T) {
	type fields struct {
		storage    storage.Storage
		collection []models.Gauge
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "when inmemory storage",
			fields: fields{
				storage:    storage.NewMemory(),
				collection: []models.Gauge{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			g := &Gauge{
				storage: tt.fields.storage,
				lg:      lg,
			}
			_, err = g.SaveCollection(context.Background(), tt.fields.collection)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestGauge_Create(t *testing.T) {
	type fields struct {
		storage storage.Storage
		record  models.Gauge
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "when inmemory storage",
			fields: fields{
				storage: storage.NewMemory(),
				record:  models.Gauge{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			g := &Gauge{
				storage: tt.fields.storage,
				lg:      lg,
			}
			_, err = g.Create(context.Background(), &tt.fields.record)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestGauge_saveCollToDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		storage *mocks.MockDBAble
		Records []models.Gauge
		lg      *logging.ZapLogger
	}
	type args struct {
		ctx  context.Context
		coll []models.Gauge
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(f *fields)
		wantErr bool
	}{
		{
			name: "successfully save collection to DB",
			args: args{
				ctx: context.Background(),
				coll: []models.Gauge{
					{Name: "test_gauge1", Value: 42.42},
					{Name: "test_gauge2", Value: 24.24},
				},
			},
			prepare: func(f *fields) {
				f.storage.EXPECT().BeginTx(
					gomock.Any(),
					gomock.Any(),
				).Return(&sql.Tx{}, nil)

				// First insert
				f.storage.EXPECT().QueryRowContext(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(nil)

				// Second insert
				f.storage.EXPECT().QueryRowContext(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(nil)

				f.storage.EXPECT().CommitTx(
					gomock.Any(),
					gomock.Any(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error when saving collection to DB",
			args: args{
				ctx: context.Background(),
				coll: []models.Gauge{
					{Name: "test_gauge1", Value: 42.42},
				},
			},
			prepare: func(f *fields) {
				f.storage.EXPECT().BeginTx(
					gomock.Any(),
					gomock.Any(),
				).Return(&sql.Tx{}, nil)

				f.storage.EXPECT().QueryRowContext(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(fmt.Errorf("database error"))

				f.storage.EXPECT().RollbackTx(
					gomock.Any(),
					gomock.Any(),
				).Return(nil)

				f.storage.EXPECT().CommitTx(
					gomock.Any(),
					gomock.Any(),
				).Return(nil)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.storage = mocks.NewMockDBAble(ctrl)
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			tt.fields.lg = lg
			c := &Gauge{
				storage: tt.fields.storage,
				Records: tt.fields.Records,
				lg:      tt.fields.lg,
			}

			tt.prepare(&tt.fields)

			got, err := c.saveCollToDB(tt.args.ctx, tt.fields.storage, tt.args.coll)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

func TestGauge_lastFromDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		storage *mocks.MockDBAble
		Records []models.Gauge
		lg      *logging.ZapLogger
	}
	type args struct {
		ctx   context.Context
		mName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(f *fields)
		wantErr bool
	}{
		{
			name: "successfully get last gauge from DB",
			args: args{
				ctx:   context.Background(),
				mName: "test_gauge",
			},
			prepare: func(f *fields) {
				f.storage.EXPECT().QueryRowContext(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(nil)
			},
		},
		{
			name: "error when getting last gauge from DB",
			args: args{
				ctx:   context.Background(),
				mName: "test_gauge",
			},
			prepare: func(f *fields) {
				f.storage.EXPECT().QueryRowContext(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(fmt.Errorf("database error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.storage = mocks.NewMockDBAble(ctrl)
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			tt.fields.lg = lg
			c := &Gauge{
				storage: tt.fields.storage,
				Records: tt.fields.Records,
				lg:      tt.fields.lg,
			}

			tt.prepare(&tt.fields)

			got, err := c.lastFromDB(tt.args.ctx, tt.fields.storage, tt.args.mName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

func TestGauge_pushToDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		storage *mocks.MockDBAble
		Records []models.Gauge
		lg      *logging.ZapLogger
	}
	type args struct {
		ctx context.Context
		rec *models.Gauge
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(f *fields)
		wantErr bool
	}{
		{
			name: "successfully push gauge to DB",
			args: args{
				ctx: context.Background(),
				rec: &models.Gauge{
					Name:  "test_gauge",
					Value: 42.42,
				},
			},
			prepare: func(f *fields) {
				f.storage.EXPECT().QueryRowContext(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error when pushing gauge to DB",
			args: args{
				ctx: context.Background(),
				rec: &models.Gauge{
					Name:  "test_gauge",
					Value: 42.42,
				},
			},
			prepare: func(f *fields) {
				f.storage.EXPECT().QueryRowContext(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(fmt.Errorf("database error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.storage = mocks.NewMockDBAble(ctrl)
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			tt.fields.lg = lg
			c := &Gauge{
				storage: tt.fields.storage,
				Records: tt.fields.Records,
				lg:      tt.fields.lg,
			}

			tt.prepare(&tt.fields)

			got, err := c.pushToDB(tt.args.ctx, tt.fields.storage, tt.args.rec)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}
