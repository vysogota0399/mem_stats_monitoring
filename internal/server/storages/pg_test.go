package storages

import (
	"context"
	"reflect"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestNewPGConnectionOpener(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	type args struct {
		lg  *logging.ZapLogger
		cfg *config.Config
	}
	tests := []struct {
		name string
		args args
		want *PGConnectionOpener
	}{
		{
			name: "create new PGConnectionOpener with valid config",
			args: args{
				lg: lg,
				cfg: &config.Config{
					DatabaseDSN: "postgres://user:pass@localhost:5432/db",
				},
			},
			want: &PGConnectionOpener{
				atpt:           10,
				maxOpenRetries: 10,
				lg:             lg,
				dbDsn:          "postgres://user:pass@localhost:5432/db",
			},
		},
		{
			name: "create new PGConnectionOpener with empty config",
			args: args{
				lg:  lg,
				cfg: &config.Config{},
			},
			want: &PGConnectionOpener{
				atpt:           10,
				maxOpenRetries: 10,
				lg:             lg,
				dbDsn:          "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPGConnectionOpener(tt.args.lg, tt.args.cfg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPGConnectionOpener() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewGooseMigrator(t *testing.T) {
	tests := []struct {
		name string
		want *GooseMigrator
	}{
		{
			name: "create new GooseMigrator",
			want: &GooseMigrator{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewGooseMigrator(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGooseMigrator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGooseMigrator_Migrate(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name: "when success",
		},
	}

	ctx := context.Background()
	cfg := &config.Config{
		DatabaseDSN: "postgres://postgres:secret@db:5432/mem_stats_monitoring_server",
	}
	pgContainer := RunTestPgContainer(t, ctx, cfg)
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	assert.NoError(t, err)

	cfg.DatabaseDSN = connStr

	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := NewPGConnectionOpener(lg, cfg)
			conn, err := op.OpenDB(ctx)
			assert.NoError(t, err)

			err = NewGooseMigrator().Migrate(conn)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPGConnectionOpener_OpenDB(t *testing.T) {
	ctx := context.Background()
	pgContainer := RunTestPgContainer(t, ctx, &config.Config{
		DatabaseDSN: "postgres://postgres:123@db:5432/mem_stats_monitoring_server",
	})
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	assert.NoError(t, err)

	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	type fields struct {
		dbDsn string
	}
	tests := []struct {
		name    string
		fields  fields
		prepare func(*postgres.PostgresContainer)
		wantErr bool
	}{
		{
			name: "when connection open error",
			fields: fields{
				dbDsn: "postgres://mysql:123@db:5432/mem_stats_monitoring_server",
			},
			wantErr: true,
		},
		{
			name: "when connection opened",
			fields: fields{
				dbDsn: connStr,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := NewPGConnectionOpener(lg, &config.Config{DatabaseDSN: tt.fields.dbDsn})
			conn, err := op.OpenDB(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, conn)
		})
	}
}

func TestNewPG(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)
	initPG(t, lg)
}

func initPG(t testing.TB, lg *logging.ZapLogger) *PG {
	ctx := context.Background()
	cfg := &config.Config{
		DatabaseDSN: "postgres://postgres:123@db:5432/mem_stats_monitoring_server",
	}
	pgContainer := RunTestPgContainer(t, ctx, cfg)
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	assert.NoError(t, err)

	cfg.DatabaseDSN = connStr

	op := NewPGConnectionOpener(lg, cfg)

	var (
		l fx.Lifecycle
		s fx.Shutdowner
	)
	app := fxtest.New(
		t,
		fx.Populate(&l, &s),
	)

	mg := NewGooseMigrator()

	db, err := NewPG(l, lg, cfg, mg, op)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	assert.NoError(t, app.Start(ctx))
	return db
}

func TestPG_CreateOrUpdate(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	type args struct {
		mType string
		mName string
		val   any
	}
	tests := []struct {
		name    string
		args    args
		prepare func(*PG) error
		wantErr bool
	}{
		{
			name: "when succeded for counter",
			args: args{
				mType: models.CounterType,
				mName: "test c",
				val:   1,
			},
			prepare: func(p *PG) error { return nil },
		},
		{
			name: "when succeded for gauge",
			args: args{
				mType: models.GaugeType,
				mName: "test g",
				val:   1,
			},
			prepare: func(p *PG) error { return nil },
		},
		{
			name: "when failed",
			args: args{
				mType: models.GaugeType,
				mName: "test g",
				val:   1,
			},
			prepare: func(p *PG) error {
				return p.db.Close()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg := initPG(t, lg)
			pg.CreateOrUpdate(context.Background(), tt.args.mType, tt.args.mName, tt.args.val)
		})
	}
}

func TestPG_GetGauge(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	type args struct {
		ctx    context.Context
		record *models.Gauge
	}
	tests := []struct {
		name    string
		args    args
		prepare func(*PG) error
		wantErr bool
	}{
		{
			name: "when gauge exists",
			args: args{
				ctx:    context.Background(),
				record: &models.Gauge{Name: "test_gauge"},
			},
			prepare: func(p *PG) error {
				return p.CreateOrUpdate(context.Background(), models.GaugeType, "test_gauge", 42.0)
			},
			wantErr: false,
		},
		{
			name: "when gauge does not exist",
			args: args{
				ctx:    context.Background(),
				record: &models.Gauge{Name: "non_existent"},
			},
			prepare: func(p *PG) error { return nil },
			wantErr: true,
		},
		{
			name: "when db is closed",
			args: args{
				ctx:    context.Background(),
				record: &models.Gauge{Name: "test_gauge"},
			},
			prepare: func(p *PG) error {
				return p.db.Close()
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg := initPG(t, lg)
			if err := tt.prepare(pg); err != nil {
				t.Fatalf("prepare failed: %v", err)
			}
			err := pg.GetGauge(tt.args.ctx, tt.args.record)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 42.0, tt.args.record.Value)
			}
		})
	}
}

func TestPG_GetCounter(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	type args struct {
		ctx    context.Context
		record *models.Counter
	}
	tests := []struct {
		name    string
		args    args
		prepare func(*PG) error
		wantErr bool
	}{
		{
			name: "when counter exists",
			args: args{
				ctx:    context.Background(),
				record: &models.Counter{Name: "test_counter"},
			},
			prepare: func(p *PG) error {
				return p.CreateOrUpdate(context.Background(), models.CounterType, "test_counter", int64(42))
			},
			wantErr: false,
		},
		{
			name: "when counter does not exist",
			args: args{
				ctx:    context.Background(),
				record: &models.Counter{Name: "non_existent"},
			},
			prepare: func(p *PG) error { return nil },
			wantErr: true,
		},
		{
			name: "when db is closed",
			args: args{
				ctx:    context.Background(),
				record: &models.Counter{Name: "test_counter"},
			},
			prepare: func(p *PG) error {
				return p.db.Close()
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg := initPG(t, lg)
			if err := tt.prepare(pg); err != nil {
				t.Fatalf("prepare failed: %v", err)
			}
			err := pg.GetCounter(tt.args.ctx, tt.args.record)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, int64(42), tt.args.record.Value)
			}
		})
	}
}

func TestPG_GetGauges(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		prepare func(*PG) error
		want    []models.Gauge
		wantErr bool
	}{
		{
			name: "when gauges exist",
			args: args{
				ctx: context.Background(),
			},
			prepare: func(p *PG) error {
				if err := p.CreateOrUpdate(context.Background(), models.GaugeType, "gauge1", 42.0); err != nil {
					return err
				}
				return p.CreateOrUpdate(context.Background(), models.GaugeType, "gauge2", 84.0)
			},
			want: []models.Gauge{
				{Name: "gauge1", Value: 42.0, ID: 1},
				{Name: "gauge2", Value: 84.0, ID: 2},
			},
			wantErr: false,
		},
		{
			name: "when no gauges exist",
			args: args{
				ctx: context.Background(),
			},
			prepare: func(p *PG) error { return nil },
			want:    []models.Gauge{},
			wantErr: false,
		},
		{
			name: "when db is closed",
			args: args{
				ctx: context.Background(),
			},
			prepare: func(p *PG) error {
				return p.db.Close()
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg := initPG(t, lg)
			if err := tt.prepare(pg); err != nil {
				t.Fatalf("prepare failed: %v", err)
			}
			got, err := pg.GetGauges(tt.args.ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.want, got)
			}
		})
	}
}

func TestPG_GetCounters(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		prepare func(*PG) error
		want    []models.Counter
		wantErr bool
	}{
		{
			name: "when counters exist",
			args: args{
				ctx: context.Background(),
			},
			prepare: func(p *PG) error {
				if err := p.CreateOrUpdate(context.Background(), models.CounterType, "counter1", int64(42)); err != nil {
					return err
				}
				return p.CreateOrUpdate(context.Background(), models.CounterType, "counter2", int64(84))
			},
			want: []models.Counter{
				{Name: "counter1", Value: 42, ID: 1},
				{Name: "counter2", Value: 84, ID: 2},
			},
			wantErr: false,
		},
		{
			name: "when no counters exist",
			args: args{
				ctx: context.Background(),
			},
			prepare: func(p *PG) error { return nil },
			want:    []models.Counter{},
			wantErr: false,
		},
		{
			name: "when db is closed",
			args: args{
				ctx: context.Background(),
			},
			prepare: func(p *PG) error {
				return p.db.Close()
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg := initPG(t, lg)
			if err := tt.prepare(pg); err != nil {
				t.Fatalf("prepare failed: %v", err)
			}
			got, err := pg.GetCounters(tt.args.ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.want, got)
			}
		})
	}
}

func TestPG_Ping(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	tests := []struct {
		name    string
		prepare func(*PG) error
		wantErr bool
	}{
		{
			name:    "when ok",
			prepare: func(p *PG) error { return nil },
		},
		{
			name:    "when connection closed",
			prepare: func(p *PG) error { return p.db.Close() },
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg := initPG(t, lg)
			if err := tt.prepare(pg); err != nil {
				t.Fatalf("prepare failed: %v", err)
			}
			err := pg.Ping(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPG_Tx(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		prepare func(*PG) (func(context.Context) error, error)
		want    []models.Counter
		wantErr bool
	}{
		{
			name: "when ok",
			args: args{
				ctx: context.Background(),
			},
			prepare: func(p *PG) (func(context.Context) error, error) {
				return func(c context.Context) error {
					return p.CreateOrUpdate(c, models.GaugeType, "name 1", 1)
				}, nil
			},
		},
		{
			name: "when failed",
			args: args{
				ctx: context.Background(),
			},
			prepare: func(p *PG) (func(context.Context) error, error) {
				return func(c context.Context) error {
					return p.GetGauge(c, &models.Gauge{})
				}, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg := initPG(t, lg)
			f, err := tt.prepare(pg)
			assert.NoError(t, err)

			err = pg.Tx(tt.args.ctx, f)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
