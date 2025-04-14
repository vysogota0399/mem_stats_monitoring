package storage

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	mocks "github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage/interfaces"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func TestDBStorage_All(t *testing.T) {
	type fields struct {
		dbDsn          string
		db             *sql.DB
		maxOpenRetries uint8
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]map[string][]string
	}{
		{
			name: "returens map of gauges",
			fields: fields{
				dbDsn:          "postgres://postgres:postgres@localhost:5432/postgres",
				db:             nil,
				maxOpenRetries: 0,
			},
			want: map[string]map[string][]string{},
		},
	}
	for _, tt := range tests {
		lg, err := logging.MustZapLogger(zap.InfoLevel)
		assert.NoError(t, err)

		t.Run(tt.name, func(t *testing.T) {
			s := &DBStorage{
				dbDsn:          tt.fields.dbDsn,
				db:             tt.fields.db,
				lg:             lg,
				maxOpenRetries: tt.fields.maxOpenRetries,
			}
			got := s.All()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDBStorage_Last(t *testing.T) {
	type fields struct {
		dbDsn          string
		db             interfaces.IDB
		maxOpenRetries uint8
	}
	type args struct {
		mType string
		mName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "returns last value",
			fields: fields{
				dbDsn:          "postgres://postgres:postgres@localhost:5432/postgres",
				db:             nil,
				maxOpenRetries: 0,
			},
			args: args{
				mType: "gauge",
				mName: "gauge1",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		lg, err := logging.MustZapLogger(zap.InfoLevel)
		assert.NoError(t, err)

		t.Run(tt.name, func(t *testing.T) {
			s := &DBStorage{
				dbDsn:          tt.fields.dbDsn,
				db:             tt.fields.db,
				lg:             lg,
				maxOpenRetries: tt.fields.maxOpenRetries,
			}
			got, err := s.Last(tt.args.mType, tt.args.mName)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDBStorage_Push(t *testing.T) {
	lg, err := logging.MustZapLogger(zap.InfoLevel)
	assert.NoError(t, err)

	type fields struct {
		dbDsn          string
		db             interfaces.IDB
		maxOpenRetries uint8
	}
	type args struct {
		mType string
		mName string
		val   any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "returns no error",
			fields: fields{
				dbDsn:          "postgres://postgres:postgres@localhost:5432/postgres",
				db:             nil,
				maxOpenRetries: 0,
			},
			args: args{
				mType: "gauge",
				mName: "gauge1",
				val:   "1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DBStorage{
				dbDsn:          tt.fields.dbDsn,
				db:             tt.fields.db,
				lg:             lg,
				maxOpenRetries: tt.fields.maxOpenRetries,
			}
			err := s.Push(tt.args.mType, tt.args.mName, tt.args.val)
			assert.NoError(t, err)
		})
	}
}

func TestDBStorage_Ping(t *testing.T) {
	type fields struct {
		dbDsn          string
		db             *mocks.MockIDB
		maxOpenRetries uint8
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		prepare func(f *fields)
	}{
		{
			name: "returns no error",
			fields: fields{
				dbDsn:          "postgres://postgres:postgres@localhost:5432/postgres",
				maxOpenRetries: 0,
			},
			prepare: func(f *fields) {
				f.db.EXPECT().Ping().Return(nil)
			},
		},
	}
	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		lg, err := logging.MustZapLogger(zap.InfoLevel)
		assert.NoError(t, err)

		t.Run(tt.name, func(t *testing.T) {
			tt.fields.db = mocks.NewMockIDB(ctrl)
			s := &DBStorage{
				lg:             lg,
				dbDsn:          tt.fields.dbDsn,
				db:             tt.fields.db,
				maxOpenRetries: tt.fields.maxOpenRetries,
			}
			tt.prepare(&tt.fields)

			err := s.Ping()
			assert.NoError(t, err)
		})
	}
}

func TestNewDBStorage(t *testing.T) {
	type args struct {
		migrator         *mocks.MockMigrator
		connectionOpener *mocks.MockConnectionOpener
		db               *mocks.MockIDB
	}
	tests := []struct {
		name    string
		prepare func(args *args)
		args    args
		wantErr bool
	}{
		{
			name: "returns new DBStorage",
			args: args{
				migrator:         &mocks.MockMigrator{},
				connectionOpener: &mocks.MockConnectionOpener{},
				db:               &mocks.MockIDB{},
			},
			prepare: func(args *args) {
				args.migrator.EXPECT().Migrate(gomock.Any()).Return(nil)
				args.connectionOpener.EXPECT().OpenDB(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "returns error when migrator failed",
			args: args{
				migrator:         &mocks.MockMigrator{},
				connectionOpener: &mocks.MockConnectionOpener{},
				db:               &mocks.MockIDB{},
			},
			prepare: func(args *args) {
				args.db.EXPECT().Close().Return(nil)
				args.connectionOpener.EXPECT().OpenDB(gomock.Any(), gomock.Any()).Return(args.db, nil)
				args.migrator.EXPECT().Migrate(gomock.Any()).Return(errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "returns error when connection opener failed",
			args: args{
				migrator:         &mocks.MockMigrator{},
				connectionOpener: &mocks.MockConnectionOpener{},
				db:               &mocks.MockIDB{},
			},
			prepare: func(args *args) {
				args.connectionOpener.EXPECT().OpenDB(gomock.Any(), gomock.Any()).Return(nil, errors.New("test: open db error"))
			},
			wantErr: true,
		},
	}

	lg, err := logging.MustZapLogger(zap.InfoLevel)
	assert.NoError(t, err)

	cfg := config.Config{
		DatabaseDSN: "postgres://postgres:postgres@localhost:5432/postgres",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.migrator = mocks.NewMockMigrator(gomock.NewController(t))
			tt.args.connectionOpener = mocks.NewMockConnectionOpener(gomock.NewController(t))
			tt.args.db = mocks.NewMockIDB(gomock.NewController(t))
			tt.prepare(&tt.args)

			_, err := NewDBStorage(context.Background(), cfg, &errgroup.Group{}, lg, tt.args.migrator, tt.args.connectionOpener)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDBStorage_QueryRowContext(t *testing.T) {
	type fields struct {
		db *mocks.MockIDB
	}
	type args struct {
		ctx    context.Context
		query  string
		args   []any
		result ResultFunc
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(f *fields)
		wantErr bool
	}{
		{
			name:   "returns no error",
			fields: fields{},
			args: args{
				ctx:    context.Background(),
				query:  "SELECT * FROM users",
				args:   []any{},
				result: func(rows *sql.Rows) error { return nil },
			},
			prepare: func(f *fields) {
				rows := &sql.Rows{}
				f.db.EXPECT().QueryContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(rows, nil)
			},
			wantErr: false,
		},
		{
			name:   "returns error when query failed",
			fields: fields{},
			args: args{
				ctx:    context.Background(),
				query:  "SELECT * FROM users",
				args:   []any{},
				result: func(rows *sql.Rows) error { return nil },
			},
			prepare: func(f *fields) {
				f.db.EXPECT().QueryContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("test: query error"))
			},
			wantErr: true,
		},
		{
			name:   "returns error when function result failed",
			fields: fields{},
			args: args{
				ctx:    context.Background(),
				query:  "SELECT * FROM users",
				args:   []any{},
				result: func(rows *sql.Rows) error { return errors.New("test: result error") },
			},
			prepare: func(f *fields) {
				rows := &sql.Rows{}
				f.db.EXPECT().QueryContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(rows, nil)
			},
			wantErr: true,
		},
	}

	lg, err := logging.MustZapLogger(zap.InfoLevel)
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tt.fields.db = mocks.NewMockIDB(ctrl)
			s := &DBStorage{
				db:       tt.fields.db,
				lg:       lg,
				testMode: true,
			}

			tt.prepare(&tt.fields)

			if tt.wantErr {
				assert.Error(t, s.QueryRowContext(tt.args.ctx, tt.args.query, tt.args.args, tt.args.result))
			} else {
				assert.NoError(t, s.QueryRowContext(tt.args.ctx, tt.args.query, tt.args.args, tt.args.result))
			}
		})
	}
}

func TestDBStorage_BeginTx(t *testing.T) {
	type fields struct {
		db *mocks.MockIDB
	}
	type args struct {
		ctx  context.Context
		opts *sql.TxOptions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *sql.Tx
		prepare func(f *fields)
		wantErr bool
	}{
		{
			name:   "returns new tx",
			fields: fields{},
			args: args{
				ctx:  context.Background(),
				opts: &sql.TxOptions{},
			},
			prepare: func(f *fields) {
				f.db.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(&sql.Tx{}, nil)
			},
			want:    &sql.Tx{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.db = mocks.NewMockIDB(gomock.NewController(t))
			s := &DBStorage{
				db: tt.fields.db,
			}
			tt.prepare(&tt.fields)
			got, err := s.BeginTx(tt.args.ctx, tt.args.opts)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}
