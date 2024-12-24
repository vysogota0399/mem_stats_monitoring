package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
)

type DataReader struct {
	file    *os.File
	mtx     *sync.Mutex
	scanner bufio.Scanner
	strg    Storage
}

func newDataReader(filename string, strg Storage) (*DataReader, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &DataReader{
		file:    f,
		scanner: *bufio.NewScanner(f),
		strg:    strg,
		mtx:     &sync.Mutex{},
	}, nil
}

type DataWriter struct {
	file *os.File
	mtx  *sync.Mutex
	strg Storage
}

func newDataWriter(filename string, strg Storage) (*DataWriter, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &DataWriter{
		file: f,
		strg: strg,
		mtx:  &sync.Mutex{},
	}, nil
}

func (dw *DataWriter) close() error {
	return dw.file.Close()
}

func (dw *DataWriter) writeMetric(m any) error {
	dw.mtx.Lock()
	defer dw.mtx.Unlock()

	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	_, err = dw.file.Write(b)
	return err
}

type PersistentMemory struct {
	Memory
	syncExport bool
	dw         *DataWriter
	dr         *DataReader
}

func NewPersistentMemory(c config.Config) (*PersistentMemory, error) {
	s := &PersistentMemory{
		Memory: *NewMemory(),
	}
	s.syncExport = c.StoreInterval == 0
	dw, err := newDataWriter(c.FileStoragePath, s)
	if err != nil {
		return nil, err
	}
	dr, err := newDataReader(c.FileStoragePath, s)
	if err != nil {
		return nil, err
	}

	s.dw = dw
	s.dr = dr

	if c.Restore {

	}

	return s, nil
}

func (m *PersistentMemory) Push(mType, mName string, val any) error {
	if err := m.Memory.Push(mType, mName, val); err != nil {
		return err
	}

	if !m.syncExport {
		return nil
	}

	return m.dw.writeMetric(val)
}

func (m *PersistentMemory) restore() error {
	return nil
}
