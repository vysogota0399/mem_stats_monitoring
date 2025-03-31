package pubsub

import (
	"bytes"
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"golang.org/x/sync/errgroup"
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

			errg, ctx := errgroup.WithContext(ctx)
			s.Start(ctx, errg)

			pb := newPublisher(lg, tt.fields.ch)
			for _, m := range tt.want.el {
				pb.Push(&m)
			}

			cancel()
			err = errg.Wait()
			assert.NoError(t, err)
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

type ReadWriteCloserBuffer struct {
	*bytes.Buffer
}

func (b ReadWriteCloserBuffer) Close() error {
	return nil
}

func TestSubscriber_appendMetrics(t *testing.T) {
	type fields struct {
		b ReadWriteCloserBuffer
	}
	type args struct {
		ctx      context.Context
		messages []*Message
	}
	type want struct {
		data     string
		hasError bool
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "when empty message do nothing",
			args: args{messages: make([]*Message, 0)},
			fields: fields{
				b: ReadWriteCloserBuffer{Buffer: &bytes.Buffer{}},
			},
			want: want{
				data: "",
			},
		},
		{
			name: "write messages to target",
			args: args{
				messages: []*Message{
					{
						MName:  "fiz",
						MType:  "baz",
						MValue: "1",
					},
				},
			},
			fields: fields{
				b: ReadWriteCloserBuffer{Buffer: &bytes.Buffer{}},
			},
			want: want{
				data: "{\"name\":\"fiz\",\"type\":\"baz\",\"value\":\"1\"}\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Subscriber{dw: newDataWriter(tt.fields.b)}
			err := s.appendMetrics(tt.args.messages...)
			assert.Equal(t, tt.want.hasError, err != nil)

			actualBuffer := &bytes.Buffer{}
			actualBuffer.Write(tt.fields.b.Bytes())

			assert.Equal(t, tt.want.data, actualBuffer.String())
		})
	}
}

func TestNewPubSub(t *testing.T) {
	lg, err := logging.MustZapLogger(-1)
	assert.NoError(t, err)

	ps := NewPubSub(lg, config.Config{}, ReadWriteCloserBuffer{Buffer: &bytes.Buffer{}})
	assert.NotNil(t, ps.Pb)
	assert.NotNil(t, ps.Sb)
	assert.NotNil(t, ps.ch)
}

func TestPublisher_Push(t *testing.T) {
	type fields struct {
		ch chan *Message
	}
	type args struct {
		m *Message
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "pus new value",
			fields: fields{ch: make(chan *Message)},
			args:   args{m: &Message{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)

			p := &Publisher{
				ch:  tt.fields.ch,
				lg:  lg,
				ctx: context.Background(),
			}
			go func() {
				p.Push(tt.args.m)
			}()

			assert.Equal(t, tt.args.m, <-p.ch)
		})
	}
}
