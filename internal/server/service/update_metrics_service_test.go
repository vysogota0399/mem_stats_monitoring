package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/mocks"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
)

func TestUpdateMetricsService_Call(t *testing.T) {
	cntr := models.Counter{Value: 1, Name: "test", ID: 1}
	gg := models.Gauge{Value: 1, Name: "test", ID: 1}

	type fields struct {
		counterRep *mocks.MockCntrRep
		gaugeRep   *mocks.MockGGRep
		ctx        context.Context
		cs         []models.Counter
		ggs        []models.Gauge
	}
	tests := []struct {
		name    string
		fields  fields
		want    *UpdateMetricsServiceResult
		prepare func(f *fields)
		wantErr bool
	}{
		{
			name: "when params contains both metrics",
			fields: fields{
				ctx: context.Background(),
				cs:  []models.Counter{cntr},
				ggs: []models.Gauge{gg},
			},
			prepare: func(f *fields) {
				f.counterRep.EXPECT().SaveCollection(f.ctx, gomock.Any()).Return(f.cs, nil)
				f.gaugeRep.EXPECT().SaveCollection(f.ctx, gomock.Any()).Return(f.ggs, nil)
			},
			wantErr: false,
			want: &UpdateMetricsServiceResult{
				cntrs: []models.Counter{cntr},
				ggs:   []models.Gauge{gg},
			},
		},
		{
			name: "when params contains only counters",
			fields: fields{
				ctx: context.Background(),
				cs:  []models.Counter{cntr},
				ggs: []models.Gauge{},
			},
			prepare: func(f *fields) {
				f.counterRep.EXPECT().SaveCollection(f.ctx, gomock.Any()).Return(f.cs, nil)
				f.gaugeRep.EXPECT().SaveCollection(f.ctx, gomock.Any()).Return(f.ggs, nil)
			},
			want: &UpdateMetricsServiceResult{
				cntrs: []models.Counter{cntr},
				ggs:   []models.Gauge{},
			},
			wantErr: false,
		},
		{
			name: "when repository returns error",
			fields: fields{
				ctx: context.Background(),
				cs:  []models.Counter{cntr},
				ggs: []models.Gauge{gg},
			},
			prepare: func(f *fields) {
				f.counterRep.EXPECT().SaveCollection(f.ctx, gomock.Any()).Return(nil, errors.New(""))
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			fields := tt.fields
			fields.counterRep = mocks.NewMockCntrRep(ctrl)
			fields.gaugeRep = mocks.NewMockGGRep(ctrl)

			s := &UpdateMetricsService{
				counterRep: fields.counterRep,
				gaugeRep:   fields.gaugeRep,
			}

			tt.prepare(&fields)

			got, err := s.Call(tt.fields.ctx, UpdateMetricsServiceParams{
				{MType: models.CounterType},
				{MType: models.GaugeType},
			})

			if tt.wantErr {
				assert.NotNil(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
