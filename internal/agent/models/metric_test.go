package models

import (
	"testing"
)

func TestMetric_String(t *testing.T) {
	type fields struct {
		Name  string
		Type  string
		Value string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "empty metric",
			fields: fields{
				Name:  "",
				Type:  "",
				Value: "",
			},
			want: `{"name":"","type":"","value":""}`,
		},
		{
			name: "gauge metric",
			fields: fields{
				Name:  "Alloc",
				Type:  GaugeType,
				Value: "123.45",
			},
			want: `{"name":"Alloc","type":"gauge","value":"123.45"}`,
		},
		{
			name: "counter metric",
			fields: fields{
				Name:  "PollCount",
				Type:  CounterType,
				Value: "42",
			},
			want: `{"name":"PollCount","type":"counter","value":"42"}`,
		},
		{
			name: "metric with special characters",
			fields: fields{
				Name:  "test-metric",
				Type:  GaugeType,
				Value: "1.2.3",
			},
			want: `{"name":"test-metric","type":"gauge","value":"1.2.3"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Metric{
				Name:  tt.fields.Name,
				Type:  tt.fields.Type,
				Value: tt.fields.Value,
			}
			if got := m.String(); got != tt.want {
				t.Errorf("Metric.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetric_FromJSON(t *testing.T) {
	type fields struct {
		Name  string
		Type  string
		Value string
	}
	type args struct {
		inp []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid gauge metric",
			fields: fields{
				Name:  "Alloc",
				Type:  GaugeType,
				Value: "123.45",
			},
			args: args{
				inp: []byte(`{"name":"Alloc","type":"gauge","value":"123.45"}`),
			},
			wantErr: false,
		},
		{
			name: "valid counter metric",
			fields: fields{
				Name:  "PollCount",
				Type:  CounterType,
				Value: "42",
			},
			args: args{
				inp: []byte(`{"name":"PollCount","type":"counter","value":"42"}`),
			},
			wantErr: false,
		},
		{
			name: "empty metric",
			fields: fields{
				Name:  "",
				Type:  "",
				Value: "",
			},
			args: args{
				inp: []byte(`{"name":"","type":"","value":""}`),
			},
			wantErr: false,
		},
		{
			name: "invalid JSON",
			fields: fields{
				Name:  "",
				Type:  "",
				Value: "",
			},
			args: args{
				inp: []byte(`{"name":"Alloc","type":"gauge","value":"123.45"`), // missing closing brace
			},
			wantErr: true,
		},
		{
			name: "missing required field",
			fields: fields{
				Name:  "",
				Type:  "",
				Value: "",
			},
			args: args{
				inp: []byte(`{"name":"Alloc","type":"gauge"}`), // missing value field
			},
			wantErr: false, // easyjson handles missing fields gracefully
		},
		{
			name: "extra fields",
			fields: fields{
				Name:  "Alloc",
				Type:  GaugeType,
				Value: "123.45",
			},
			args: args{
				inp: []byte(`{"name":"Alloc","type":"gauge","value":"123.45","extra":"field"}`),
			},
			wantErr: false, // extra fields are ignored
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				Name:  tt.fields.Name,
				Type:  tt.fields.Type,
				Value: tt.fields.Value,
			}
			if err := m.FromJSON(tt.args.inp); (err != nil) != tt.wantErr {
				t.Errorf("Metric.FromJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
