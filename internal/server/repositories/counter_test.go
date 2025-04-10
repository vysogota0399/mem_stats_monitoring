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

func TestCreate(t *testing.T) {
	tasks := []struct {
		record      models.Counter
		wantsRecord models.Counter
		name        string
		storage     storage.Storage
	}{
		{
			record:      models.Counter{Value: 1, Name: "test"},
			wantsRecord: models.Counter{Value: 1, Name: "test"},
			name:        "when storage empty create then record to state",
			storage: storage.NewMemStorageWithData(
				map[string]map[string][]string{},
			),
		},
		{
			record:      models.Counter{Value: 2, Name: "test"},
			wantsRecord: models.Counter{Value: 3, Name: "test"},
			name:        "when storage has record then create add one",
			storage: storage.NewMemStorageWithData(
				map[string]map[string][]string{"counter": {"test": []string{`{"value": 1, "name": "test"}`}}},
			),
		},
	}

	for _, tt := range tasks {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			subject := NewCounter(tt.storage, lg)
			record, err := subject.Create(context.Background(), &tt.record)
			assert.NoError(t, err)
			assert.NotNil(t, record)
			assert.Equal(t, tt.wantsRecord, *record)
		})
	}
}

func TestLast(t *testing.T) {
	tasks := []struct {
		wantsRecord *models.Counter
		wantsError  error
		name        string
		storage     storage.Storage
		searchName  string
	}{
		{
			searchName:  "test",
			wantsRecord: &models.Counter{Value: 1, Name: "test"},
			wantsError:  nil,
			name:        "when storage has record then returns error",
			storage: storage.NewMemStorageWithData(
				map[string]map[string][]string{"counter": {"test": []string{`{"value": 1, "name": "test"}`}}},
			),
		},
		{
			searchName:  "test2",
			wantsRecord: nil,
			wantsError:  storage.ErrNoRecords,
			storage: storage.NewMemStorageWithData(
				map[string]map[string][]string{"counter": {"test": []string{`{"value": 1, "name": "test"}`}}},
			),
		},
	}

	for _, tt := range tasks {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			subject := NewCounter(tt.storage, lg)
			record, err := subject.Last(context.Background(), tt.searchName)
			assert.Equal(t, record, tt.wantsRecord)
			assert.Equal(t, err, tt.wantsError)
		})
	}
}

func TestCounter_All(t *testing.T) {
	type fields struct {
		storage storage.Storage
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string][]models.Counter
	}{
		{
			name: "when storage has values then returns collection",
			fields: fields{
				storage: storage.NewMemStorageWithData(
					map[string]map[string][]string{
						"counter": {
							"fiz": []string{`{"value": 0, "name": "fiz"}`},
							"baz": []string{`{"value": 0, "name": "baz"}`},
						},
					},
				),
			},
			want: map[string][]models.Counter{
				"fiz": {models.Counter{Name: "fiz", Value: 0}},
				"baz": {models.Counter{Name: "baz", Value: 0}},
			},
		},
		{
			name: "when storage has no values then returns empty collection",
			fields: fields{
				storage: storage.NewMemStorageWithData(
					map[string]map[string][]string{},
				),
			},
			want: map[string][]models.Counter{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			c := NewCounter(tt.fields.storage, lg)
			actual := c.All()
			assert.Equal(t, tt.want, actual)
		})
	}
}

func TestCounter_SaveCollection(t *testing.T) {
	type fields struct {
		storage    storage.Storage
		collection []models.Counter
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
				collection: []models.Counter{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			g := &Counter{
				storage: tt.fields.storage,
				lg:      lg,
			}
			_, err = g.SaveCollection(context.Background(), tt.fields.collection)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestCounter_Create(t *testing.T) {
	type fields struct {
		storage storage.Storage
		record  models.Counter
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
				record:  models.Counter{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			g := &Counter{
				storage: tt.fields.storage,
				lg:      lg,
			}
			_, err = g.Create(context.Background(), &tt.fields.record)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestCounter_SearchByName(t *testing.T) {
	type fields struct {
		storage storage.Storage
	}
	type args struct {
		names []string
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]models.Counter
		args   args
	}{
		{
			name: "when storage has values then returns collection",
			args: args{names: []string{"fiz"}},
			fields: fields{
				storage: storage.NewMemStorageWithData(
					map[string]map[string][]string{
						"counter": {
							"fiz": []string{`{"value": 0, "name": "fiz"}`},
							"baz": []string{`{"value": 0, "name": "baz"}`},
						},
					},
				),
			},
			want: map[string]models.Counter{
				"fiz": {Name: "fiz", Value: 0},
			},
		},
		{
			name: "when storage has no values then returns empty collection",
			fields: fields{
				storage: storage.NewMemStorageWithData(
					map[string]map[string][]string{},
				),
			},
			want: map[string]models.Counter{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			c := &Counter{
				storage: tt.fields.storage,
				lg:      lg,
			}
			got, err := c.SearchByName(context.Background(), tt.args.names)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCounter_pushToDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		storage *mocks.MockDBAble
		Records []models.Counter
		lg      *logging.ZapLogger
	}
	type args struct {
		ctx context.Context
		rec *models.Counter
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(f *fields)
		want    *models.Counter
		wantErr bool
	}{
		{
			name: "successfully push counter to DB",
			args: args{
				ctx: context.Background(),
				rec: &models.Counter{
					Name:  "test_counter",
					Value: 42,
				},
			},
			prepare: func(f *fields) {
				f.storage.EXPECT().QueryRowContext(
					gomock.Any(),
					gomock.Eq(`
			INSERT INTO counters(name, value)
			VALUES ($1, $2)
			RETURNING id
		`),
					gomock.Eq([]any{"test_counter", int64(42)}),
					gomock.Any(),
				).Return(nil)
			},
			want: &models.Counter{
				Name:  "test_counter",
				Value: 42,
			},
			wantErr: false,
		},
		{
			name: "error when pushing counter to DB",
			args: args{
				ctx: context.Background(),
				rec: &models.Counter{
					Name:  "test_counter",
					Value: 42,
				},
			},
			prepare: func(f *fields) {
				f.storage.EXPECT().QueryRowContext(
					gomock.Any(),
					gomock.Eq(`
			INSERT INTO counters(name, value)
			VALUES ($1, $2)
			RETURNING id
		`),
					gomock.Eq([]any{"test_counter", int64(42)}),
					gomock.Any(),
				).Return(fmt.Errorf("database error"))
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.storage = mocks.NewMockDBAble(ctrl)
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			tt.fields.lg = lg
			c := &Counter{
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
				assert.Equal(t, tt.want.Name, got.Name)
				assert.Equal(t, tt.want.Value, got.Value)
			}
		})
	}
}

func TestCounter_lastFromDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		storage *mocks.MockDBAble
		Records []models.Counter
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
		want    *models.Counter
		wantErr bool
	}{
		{
			name: "successfully get last counter from DB",
			args: args{
				ctx:   context.Background(),
				mName: "test_counter",
			},
			prepare: func(f *fields) {
				f.storage.EXPECT().QueryRowContext(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(nil)
			},
			want: &models.Counter{
				Name: "test_counter",
			},
			wantErr: false,
		},
		{
			name: "error when getting last counter from DB",
			args: args{
				ctx:   context.Background(),
				mName: "test_counter",
			},
			prepare: func(f *fields) {
				f.storage.EXPECT().QueryRowContext(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(fmt.Errorf("database error"))
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.storage = mocks.NewMockDBAble(ctrl)
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			tt.fields.lg = lg
			c := &Counter{
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
				assert.Equal(t, tt.want.Name, got.Name)
				assert.Equal(t, tt.want.Value, got.Value)
			}
		})
	}
}

func TestCounter_saveCollToDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		storage *mocks.MockDBAble
		Records []models.Counter
		lg      *logging.ZapLogger
	}
	type args struct {
		ctx  context.Context
		coll []models.Counter
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(f *fields)
		want    []models.Counter
		wantErr bool
	}{
		{
			name: "successfully save collection to DB",
			args: args{
				ctx: context.Background(),
				coll: []models.Counter{
					{Name: "test_counter1", Value: 42},
					{Name: "test_counter2", Value: 24},
				},
			},
			prepare: func(f *fields) {
				// First SearchByName call
				f.storage.EXPECT().QueryRowContext(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(nil)

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
			want: []models.Counter{
				{Name: "test_counter1", Value: 42},
				{Name: "test_counter2", Value: 24},
			},
			wantErr: false,
		},
		{
			name: "error when saving collection to DB",
			args: args{
				ctx: context.Background(),
				coll: []models.Counter{
					{Name: "test_counter1", Value: 42},
				},
			},
			prepare: func(f *fields) {
				// First SearchByName call
				f.storage.EXPECT().QueryRowContext(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(nil)

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
			want:    nil,
			wantErr: true,
		},
		{
			name: "error when searching by name failed	",
			args: args{
				ctx: context.Background(),
				coll: []models.Counter{
					{Name: "test_counter1", Value: 42},
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
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.storage = mocks.NewMockDBAble(ctrl)
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			tt.fields.lg = lg
			c := &Counter{
				storage: tt.fields.storage,
				Records: tt.fields.Records,
				lg:      tt.fields.lg,
			}

			tt.prepare(&tt.fields)

			got, err := c.saveCollToDB(tt.args.ctx, tt.fields.storage, tt.args.coll)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
