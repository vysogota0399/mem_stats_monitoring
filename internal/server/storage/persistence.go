package storage

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/logger"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
)

type DataReader struct {
	file    *os.File
	mtx     *sync.Mutex
	scanner bufio.Scanner
	strg    Storage
	ctx     context.Context
}

func newDataReader(ctx context.Context, filename string, strg Storage) (*DataReader, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &DataReader{
		file:    f,
		scanner: *bufio.NewScanner(f),
		strg:    strg,
		mtx:     &sync.Mutex{},
		ctx:     ctx,
	}, nil
}

func (dr *DataReader) Close() error {
	return dr.file.Close()
}

type DataWriter struct {
	mtx    *sync.Mutex
	strg   Storage
	ctx    context.Context
	file   *os.File
	writer *bufio.Writer
}

func newDataWriter(ctx context.Context, filename string, strg Storage) (*DataWriter, error) {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &DataWriter{
		strg:   strg,
		file:   f,
		mtx:    &sync.Mutex{},
		ctx:    ctx,
		writer: bufio.NewWriter(f),
	}, nil
}

func (dw *DataWriter) Close() error {
	return dw.file.Close()
}

func (dw *DataWriter) appendMetric(m any) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	if _, err = dw.writer.Write(b); err != nil {
		return err
	}

	if err := dw.writer.WriteByte('\n'); err != nil {
		return err
	}

	return dw.writer.Flush()
}

func (dw *DataWriter) Write(b []byte) (n int, err error) {
	n, err = dw.file.Write(b)
	return
}

func (dw *DataWriter) Truncate() (err error) {
	_, err = dw.file.Seek(0, io.SeekStart)

	err = dw.file.Truncate(0)
	return
}

type fileDataDumper interface {
	Dump() error
}

type PMScheduller struct {
	ctx  context.Context
	dur  time.Duration
	strg fileDataDumper
}

func NewPMScheduller(ctx context.Context, dur int64, strg fileDataDumper) *PMScheduller {
	return &PMScheduller{
		ctx:  ctx,
		dur:  time.Second * time.Duration(dur),
		strg: strg,
	}
}

func (s *PMScheduller) Start() {
	for {
		select {
		case <-s.ctx.Done():
			logger.Log.Sugar().Debugln("Gracefull shutdown PMScheduller")
			return
		case <-time.After(s.dur):
			if err := s.strg.Dump(); err != nil {
				logger.Log.Sugar().Errorf("Dump creation failed error %w", err)
				continue
			}
		}
	}
}

type PersistentMemory struct {
	Memory
	syncExport   bool
	ctx          context.Context
	fStoragePath string
	scheduller   *PMScheduller
}

func NewPersistentMemory(ctx context.Context, c config.Config, wg *sync.WaitGroup) (*PersistentMemory, error) {
	s := &PersistentMemory{
		Memory:       *NewMemory(),
		ctx:          ctx,
		fStoragePath: c.FileStoragePath,
	}
	s.syncExport = c.StoreInterval == 0
	if c.Restore {
		go s.restore()
	}

	if !s.syncExport {
		s.scheduller = NewPMScheduller(ctx, c.StoreInterval, s)
		go s.scheduller.Start()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				logger.Log.Sugar().Debugln("Gracefull shutdown PersistentMemory - PENDING")
				s.Dump()
				logger.Log.Sugar().Debugln("Gracefull shutdown PersistentMemory - OK")
				return
			}
		}
	}()

	return s, nil
}

type persistentMetric struct {
	MType  string `json:"type"`
	MName  string `json:"name"`
	MValue any    `json:"value"`
}

func (m *PersistentMemory) Push(mType, mName string, val any) error {
	dw, err := newDataWriter(m.ctx, m.fStoragePath, m)
	if err != nil {
		return err
	}
	defer dw.Close()

	if err := m.Memory.Push(mType, mName, val); err != nil {
		return err
	}

	if !m.syncExport {
		return nil
	}

	metric := persistentMetric{}

	switch val.(type) {
	case models.Counter:
		counter := val.(models.Counter)
		metric.MName = counter.Name
		metric.MType = models.CounterType
		metric.MValue = counter.Value
	case models.Gauge:
		gauge := val.(models.Gauge)
		metric.MName = gauge.Name
		metric.MType = models.GaugeType
		metric.MValue = gauge.Value
	}

	return dw.appendMetric(metric)
}

func (m *PersistentMemory) restore() error {
	dr, err := newDataReader(m.ctx, m.fStoragePath, m)
	if err != nil {
		return err
	}
	defer dr.Close()

	scanner := dr.scanner
	var sucCntr int64
	var failCntr int64

	for scanner.Scan() {
		sucCntr++
		var record persistentMetric
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			failCntr++
			logger.Log.Sugar().Errorf("Unmarshal failed record %s error %v", scanner.Text(), err)
			continue
		}

		if err := m.Memory.Push(record.MType, record.MName, record); err != nil {
			failCntr++
			logger.Log.Sugar().Errorf("Push record %v error %v", record, err)
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	logger.Log.Sugar().Debugf("Storage restored, succeded: %d, failed: %d", sucCntr, failCntr)
	return nil
}

func (m *PersistentMemory) Dump() error {
	dw, err := newDataWriter(m.ctx, m.fStoragePath, m)
	if err != nil {
		return err
	}
	defer dw.Close()

	if err := dw.Truncate(); err != nil {
		return err
	}

	uuid := uuid.NewV4()
	var sucCntr int64
	var failCntr int64
	logger.Log.Sugar().Infow("Dump start",
		"dump_uuid", uuid.String())

	records := m.Memory.All()
	logger.Log.Sugar().Debugf("All records: %v", records)

	for mtype, mnames := range records {
		for _, metrics := range mnames {
			for _, metric := range metrics {
				record := persistentMetric{}

				if err := json.Unmarshal([]byte(metric), &record); err != nil {
					failCntr++
					logger.Log.Sugar().Errorw(
						fmt.Sprintf("Dump for value %s failed with error: %w", metric, err),
						"dump_uuid", uuid.String(),
					)
					continue
				}

				switch mtype {
				case models.CounterType:
					record.MType = models.CounterType
				case models.GaugeType:
					record.MType = models.GaugeType
				}

				b := bytes.Buffer{}

				if err := json.NewEncoder(&b).Encode(record); err != nil {
					failCntr++
					logger.Log.Sugar().Errorw(
						fmt.Sprintf("Encode for value %v failed with error: %w", record, err),
						"dump_uuid", uuid.String(),
					)
					continue
				}

				if _, err := dw.Write(b.Bytes()); err != nil {
					failCntr++
					logger.Log.Sugar().Errorw(
						fmt.Sprintf("Save to file value %s failed with error: %w", b.String(), err),
						"dump_uuid", uuid.String(),
					)
					continue
				}

				sucCntr++
			}

		}
	}

	logger.Log.Sugar().Infow("Dump finished",
		"dump_uuid", uuid.String(),
		"succeded", sucCntr,
		"failed", failCntr,
	)
	return nil
}
