package clients

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
	mocks "github.com/vysogota0399/mem_stats_monitoring/internal/mocks/agent/clients"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

func BenchmarkReporter_UpdateMetrics(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	client := NewMockRequester(ctrl)
	ips := mocks.NewMockIRealIPHeaderSetter(ctrl)

	ips.EXPECT().Call(gomock.Any()).AnyTimes().Return(nil)

	response := &http.Response{StatusCode: http.StatusOK}
	response.Body = io.NopCloser(bytes.NewBuffer([]byte{}))

	lg, err := logging.MustZapLogger(&config.Config{LogLevel: 2})
	assert.NoError(b, err)

	reporter := NewCompReporter(
		"",
		lg,
		&config.Config{RateLimit: 10},
		client,
		ips,
		agent.NewMetricsRepository(storage.NewMemoryStorage(nil)),
	)
	ctx := context.Background()

	types := []string{models.CounterType, models.CounterType}
	mCount := 10_000
	metrics := make([]*models.Metric, mCount)

	for b.Loop() {
		for i := range mCount {
			metrics[i] = &models.Metric{
				Name:  gofakeit.Animal(),
				Type:  types[rand.Int31n(2)],
				Value: gofakeit.Digit(),
			}
		}

		client.EXPECT().Request(gomock.Any()).Return(response, nil)
		assert.NoError(b, reporter.UpdateMetrics(ctx, metrics))
	}
}

func BenchmarkReporter_bytesreader(b *testing.B) {
	metrics := []*models.Metric{
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
	}
	var body bytes.Buffer

	if err := json.NewEncoder(&body).Encode(metrics); err != nil {
		assert.NoError(b, err)
	}

	b.Run("when copy", func(b *testing.B) {
		for b.Loop() {
			buff := bytes.Buffer{}
			_, err := io.Copy(&buff, &body)
			assert.NoError(b, err)
		}
	})

	b.Run("when read all", func(b *testing.B) {
		for b.Loop() {
			_, err := io.ReadAll(&body)
			assert.NoError(b, err)
		}
	})
}

func TestNewReporter(t *testing.T) {
	// Create test dependencies
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: 1})
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockRequester(ctrl)
	address := "http://test-server"

	// Test the constructor
	reporter := NewReporter(address, lg, mockClient)

	// Verify all fields are properly initialized
	assert.NotNil(t, reporter)
	assert.Equal(t, address, reporter.address)
	assert.Equal(t, lg, reporter.lg)
	assert.Equal(t, mockClient, reporter.client)
	assert.Equal(t, uint8(2), reporter.maxAttempts)
	assert.Nil(t, reporter.compressor)
	assert.Nil(t, reporter.secretKey)
	assert.Nil(t, reporter.semaphore)
}

func TestReporter_UpdateMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lg, err := logging.MustZapLogger(&config.Config{LogLevel: 1})
	assert.NoError(t, err)

	type fields struct {
		client      *MockRequester
		address     string
		lg          *logging.ZapLogger
		compressor  compressor
		encryptor   *mocks.MockEncryptor
		ips         *mocks.MockIRealIPHeaderSetter
		maxAttempts uint8
		secretKey   []byte
	}
	type args struct {
		ctx   context.Context
		mType string
		mName string
		value string
	}

	type testCase struct {
		name    string
		fields  *fields
		args    args
		wantErr bool
		prepare func(*fields)
	}

	tests := []testCase{
		{
			name: "successful gauge metric update",
			fields: &fields{
				client:      NewMockRequester(ctrl),
				address:     "http://test-server",
				lg:          lg,
				compressor:  nil,
				maxAttempts: 5,
				secretKey:   nil,
				encryptor:   mocks.NewMockEncryptor(ctrl),
			},
			args: args{
				ctx:   context.Background(),
				mType: models.GaugeType,
				mName: "testGauge",
				value: "123.45",
			},
			wantErr: false,
			prepare: func(f *fields) {
				f.ips.EXPECT().Call(gomock.Any()).Return(nil)

				f.encryptor.EXPECT().
					Encrypt(gomock.Any()).
					Return("encrypted", nil)

				f.client.EXPECT().
					Request(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBuffer(nil)),
					}, nil)
			},
		},
		{
			name: "successful counter metric update",
			fields: &fields{
				client:      NewMockRequester(ctrl),
				address:     "http://test-server",
				lg:          lg,
				compressor:  nil,
				maxAttempts: 5,
				secretKey:   nil,
				encryptor:   mocks.NewMockEncryptor(ctrl),
			},
			args: args{
				ctx:   context.Background(),
				mType: models.CounterType,
				mName: "testCounter",
				value: "42",
			},
			wantErr: false,
			prepare: func(f *fields) {
				f.ips.EXPECT().Call(gomock.Any()).Return(nil)

				f.encryptor.EXPECT().
					Encrypt(gomock.Any()).
					Return("encrypted", nil)

				f.client.EXPECT().
					Request(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBuffer(nil)),
					}, nil)
			},
		},
		{
			name: "invalid metric type",
			fields: &fields{
				client:      NewMockRequester(ctrl),
				address:     "http://test-server",
				lg:          lg,
				compressor:  nil,
				maxAttempts: 5,
				secretKey:   nil,
				encryptor:   mocks.NewMockEncryptor(ctrl),
			},
			args: args{
				ctx:   context.Background(),
				mType: "invalid",
				mName: "testInvalid",
				value: "123",
			},
			wantErr: true,
			prepare: func(f *fields) {
				// No mock expectations needed as error occurs before request
			},
		},
		{
			name: "client request error",
			fields: &fields{
				client:      NewMockRequester(ctrl),
				address:     "http://test-server",
				lg:          lg,
				compressor:  nil,
				secretKey:   nil,
				maxAttempts: 1,
				encryptor:   mocks.NewMockEncryptor(ctrl),
			},
			args: args{
				ctx:   context.Background(),
				mType: models.GaugeType,
				mName: "testError",
				value: "123",
			},
			wantErr: true,
			prepare: func(f *fields) {
				f.ips.EXPECT().Call(gomock.Any()).Return(nil)

				f.encryptor.EXPECT().
					Encrypt(gomock.Any()).
					Return("encrypted", nil)

				f.client.EXPECT().Request(gomock.Any()).Return(nil, errors.New("request failed"))
			},
		},
		{
			name: "unsuccessful response status",
			fields: &fields{
				client:      NewMockRequester(ctrl),
				address:     "http://test-server",
				lg:          lg,
				compressor:  nil,
				maxAttempts: 5,
				encryptor:   mocks.NewMockEncryptor(ctrl),
				secretKey:   nil,
			},
			args: args{
				ctx:   context.Background(),
				mType: models.GaugeType,
				mName: "testBadStatus",
				value: "123",
			},
			wantErr: true,
			prepare: func(f *fields) {
				f.ips.EXPECT().Call(gomock.Any()).Return(nil)

				f.encryptor.EXPECT().
					Encrypt(gomock.Any()).
					Return("encrypted", nil)

				f.client.EXPECT().
					Request(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       io.NopCloser(bytes.NewBuffer(nil)),
					}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.ips = mocks.NewMockIRealIPHeaderSetter(ctrl)
			// Prepare mock behavior
			tt.prepare(tt.fields)

			c := NewCompReporter(
				tt.fields.address,
				tt.fields.lg,
				&config.Config{
					RateLimit:   10,
					MaxAttempts: tt.fields.maxAttempts,
					HTTPCert:    bytes.NewBuffer([]byte{}),
				},
				tt.fields.client,
				tt.fields.ips,
				agent.NewMetricsRepository(storage.NewMemoryStorage(nil)),
			)

			c.encryptor = tt.fields.encryptor
			err := c.UpdateMetric(tt.args.ctx, tt.args.mType, tt.args.mName, tt.args.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReporter_UpdateMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lg, err := logging.MustZapLogger(&config.Config{LogLevel: 1})
	assert.NoError(t, err)

	type fields struct {
		client      *MockRequester
		address     string
		ips         *mocks.MockIRealIPHeaderSetter
		lg          *logging.ZapLogger
		compressor  compressor
		maxAttempts uint8
		secretKey   []byte
	}
	type args struct {
		ctx  context.Context
		data []*models.Metric
	}

	type testCase struct {
		name    string
		fields  *fields
		args    args
		wantErr bool
		prepare func(*fields)
	}

	tests := []testCase{
		{
			name: "successful batch update with mixed metrics",
			fields: &fields{
				client:      NewMockRequester(ctrl),
				address:     "http://test-server",
				lg:          lg,
				compressor:  nil,
				maxAttempts: 5,
				secretKey:   nil,
			},
			args: args{
				ctx: context.Background(),
				data: []*models.Metric{
					{Name: "gauge1", Type: models.GaugeType, Value: "123.45"},
					{Name: "counter1", Type: models.CounterType, Value: "42"},
				},
			},
			wantErr: false,
			prepare: func(f *fields) {
				f.ips.EXPECT().Call(gomock.Any()).Return(nil)

				f.client.EXPECT().
					Request(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBuffer(nil)),
					}, nil)
			},
		},
		{
			name: "empty metrics batch",
			fields: &fields{
				client:      NewMockRequester(ctrl),
				address:     "http://test-server",
				lg:          lg,
				compressor:  nil,
				maxAttempts: 5,
				secretKey:   nil,
			},
			args: args{
				ctx:  context.Background(),
				data: []*models.Metric{},
			},
			wantErr: false,
			prepare: func(f *fields) {
				f.ips.EXPECT().Call(gomock.Any()).Return(nil)

				f.client.EXPECT().
					Request(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBuffer(nil)),
					}, nil)
			},
		},
		{
			name: "batch with invalid metric type",
			fields: &fields{
				client:      NewMockRequester(ctrl),
				address:     "http://test-server",
				lg:          lg,
				compressor:  nil,
				maxAttempts: 5,
				secretKey:   nil,
			},
			args: args{
				ctx: context.Background(),
				data: []*models.Metric{
					{Name: "valid", Type: models.GaugeType, Value: "123.45"},
					{Name: "invalid", Type: "invalid_type", Value: "42"},
				},
			},
			wantErr: true,
			prepare: func(f *fields) {
				// No expectations needed as validation should fail before request
			},
		},
		{
			name: "client request error",
			fields: &fields{
				client:      NewMockRequester(ctrl),
				address:     "http://test-server",
				lg:          lg,
				compressor:  nil,
				maxAttempts: 1,
				secretKey:   nil,
			},
			args: args{
				ctx: context.Background(),
				data: []*models.Metric{
					{Name: "metric1", Type: models.GaugeType, Value: "123.45"},
				},
			},
			wantErr: true,
			prepare: func(f *fields) {
				f.ips.EXPECT().Call(gomock.Any()).Return(nil)
				f.client.EXPECT().Request(gomock.Any()).Return(nil, errors.New("request failed"))
			},
		},
		{
			name: "unsuccessful response status",
			fields: &fields{
				client:      NewMockRequester(ctrl),
				address:     "http://test-server",
				lg:          lg,
				compressor:  nil,
				maxAttempts: 5,
				secretKey:   nil,
			},
			args: args{
				ctx: context.Background(),
				data: []*models.Metric{
					{Name: "metric1", Type: models.GaugeType, Value: "123.45"},
				},
			},
			wantErr: true,
			prepare: func(f *fields) {
				f.ips.EXPECT().Call(gomock.Any()).Return(nil)
				f.client.EXPECT().
					Request(gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       io.NopCloser(bytes.NewBuffer(nil)),
					}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.ips = mocks.NewMockIRealIPHeaderSetter(ctrl)
			// Prepare mock behavior
			tt.prepare(tt.fields)

			c := NewCompReporter(
				tt.fields.address,
				tt.fields.lg,
				&config.Config{
					RateLimit:   10,
					MaxAttempts: tt.fields.maxAttempts,
				},
				tt.fields.client,
				tt.fields.ips,
				agent.NewMetricsRepository(storage.NewMemoryStorage(nil)),
			)
			c.encryptor = nil
			err := c.UpdateMetrics(tt.args.ctx, tt.args.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_attemptDelay(t *testing.T) {
	type args struct {
		i uint8
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{
			name: "attempt delay",
			args: args{i: 0},
			want: 1 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := attemptDelay(tt.args.i); got != tt.want {
				t.Errorf("attemptDelay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_gzbody(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty input",
			input:   "",
			wantErr: false,
		},
		{
			name:    "simple text",
			input:   "Hello, World!",
			wantErr: false,
		},
		{
			name:    "long text",
			input:   strings.Repeat("This is a test string that will be compressed. ", 100),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := bytes.NewBufferString(tt.input)
			got, err := gzbody(input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
		})
	}
}

func TestReporter_signRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lg, err := logging.MustZapLogger(&config.Config{LogLevel: 1})
	assert.NoError(t, err)

	type fields struct {
		client          Requester
		address         string
		lg              *logging.ZapLogger
		compressor      compressor
		maxAttempts     uint8
		secretKey       []byte
		semaphore       *semaphore
		publicKeyPath   io.Reader
		encryptor       Encryptor
		repository      *agent.MetricsRepository
		ipAddressSetter IRealIPHeaderSetter
	}
	type args struct {
		ctx context.Context
		r   *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		setup   func(*fields, *args)
	}{
		{
			name: "no secret key - no signing",
			fields: fields{
				lg:        lg,
				secretKey: nil,
			},
			args: args{
				ctx: context.Background(),
				r:   &http.Request{},
			},
			wantErr: false,
			setup: func(f *fields, a *args) {
				// No setup needed
			},
		},
		{
			name: "missing body in context",
			fields: fields{
				lg:        lg,
				secretKey: []byte("test-key"),
			},
			args: args{
				ctx: context.Background(),
				r:   &http.Request{},
			},
			wantErr: true,
			setup: func(f *fields, a *args) {
				// No setup needed - context is empty
			},
		},
		{
			name: "successful signing",
			fields: fields{
				lg:        lg,
				secretKey: []byte("test-key"),
			},
			args: args{
				ctx: context.Background(),
				r:   &http.Request{Header: http.Header{}},
			},
			wantErr: false,
			setup: func(f *fields, a *args) {
				// Add body to context
				a.ctx = context.WithValue(a.ctx, bodyKey("body"), []byte("test-body"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(&tt.fields, &tt.args)
			}

			c := &Reporter{
				client:          tt.fields.client,
				address:         tt.fields.address,
				lg:              tt.fields.lg,
				compressor:      tt.fields.compressor,
				maxAttempts:     tt.fields.maxAttempts,
				secretKey:       tt.fields.secretKey,
				semaphore:       tt.fields.semaphore,
				publicKeyPath:   tt.fields.publicKeyPath,
				encryptor:       tt.fields.encryptor,
				repository:      tt.fields.repository,
				ipAddressSetter: tt.fields.ipAddressSetter,
			}
			err := c.signRequest(tt.args.ctx, tt.args.r)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.name == "missing body in context" {
					assert.Equal(t, ErrBodyKeyNotFoundInContext, err)
				}
			} else {
				assert.NoError(t, err)
				if tt.name == "successful signing" {
					signature := tt.args.r.Header.Get(signHeaderKey)
					assert.NotEmpty(t, signature)

					_, err := base64.StdEncoding.DecodeString(signature)
					assert.NoError(t, err)
				}
			}
		})
	}
}
