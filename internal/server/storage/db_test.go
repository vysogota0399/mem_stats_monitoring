package storage

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
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
		db             IDB
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
		db             IDB
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
		db             IDB
		maxOpenRetries uint8
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "returns no error",
			fields: fields{
				dbDsn:          "postgres://postgres:postgres@localhost:5432/postgres",
				db:             nil,
				maxOpenRetries: 0,
			},
		},
	}
	for _, tt := range tests {
		lg, err := logging.MustZapLogger(zap.InfoLevel)
		assert.NoError(t, err)

		t.Run(tt.name, func(t *testing.T) {
			s := &DBStorage{
				lg:             lg,
				dbDsn:          tt.fields.dbDsn,
				db:             tt.fields.db,
				maxOpenRetries: tt.fields.maxOpenRetries,
			}
			err := s.Ping()
			assert.NoError(t, err)
		})
	}
}
