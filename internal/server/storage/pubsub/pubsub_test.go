package pubsub

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

func TestSubscriber_Start(t *testing.T) {
	type fields struct {
		fStoragePath  string
		ch            chan *Message
		storeInterval time.Duration
		mtx           *sync.Mutex
	}

	type want struct {
		el []Message
	}

	tests := []struct {
		name   string
		fields fields
		want
	}{
		{
			name: "when async queue",
			fields: fields{
				fStoragePath:  "empty",
				ch:            make(chan *Message),
				storeInterval: time.Duration(10) * time.Minute,
				mtx:           &sync.Mutex{},
			},
			want: want{
				el: []Message{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Millisecond)
			q := newQueue()
			s := &Subscriber{
				fStoragePath:  tt.fields.fStoragePath,
				ch:            tt.fields.ch,
				q:             q,
				storeInterval: tt.fields.storeInterval,
				mtx:           tt.fields.mtx,
				lg:            lg,
				dw:            newDataWriter(os.Stderr),
			}

			wg := sync.WaitGroup{}
			s.Start(ctx, &wg)

			pb := newPublisher(lg, tt.fields.ch)
			for _, m := range tt.want.el {
				pb.Push(&m)
			}

			cancel()
			wg.Wait()
			assert.Equal(t, len(tt.want.el), q.l.Len())
		})
	}
}

func TestQueue_push(t *testing.T) {
	type args struct {
		action mAction
		m      []*Message
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "add new element",
			args: args{
				action: push,
				m: []*Message{
					{MValue: "1"},
					{MValue: "2"},
					{MValue: "3"},
				},
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := newQueue()
			for _, v := range tt.args.m {
				q.push(v)
			}

			assert.Equal(t, tt.want, q.len())
		})
	}
}

func TestQueue_pop(t *testing.T) {
	type args struct {
		action mAction
		m      []*Message
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "pop currect elements",
			args: args{
				action: push,
				m: []*Message{
					{MValue: "1"},
					{MValue: "2"},
					{MValue: "3"},
				},
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := newQueue()
			for _, v := range tt.args.m {
				q.push(v)
			}

			assert.Equal(t, *tt.args.m[0], q.pop().Value.(Message))
			assert.Equal(t, *tt.args.m[1], q.pop().Value.(Message))
			assert.Equal(t, *tt.args.m[2], q.pop().Value.(Message))
		})
	}
}
