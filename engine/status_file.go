package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const StatusFileName = "status.json"

const defaultStatusFileWriteInterval = 3 * time.Second

type statusFileWriter struct {
	path     string
	snapshot func() StatusSnapshot

	writeInterval time.Duration

	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
	doneCh  chan struct{}
}

func newStatusFileWriter(path string, snapshot func() StatusSnapshot) *statusFileWriter {
	return &statusFileWriter{
		path:          path,
		snapshot:      snapshot,
		writeInterval: defaultStatusFileWriteInterval,
	}
}

func (w *statusFileWriter) Start() {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})
	w.running = true
	w.stopCh = stopCh
	w.doneCh = doneCh
	writeInterval := w.writeInterval
	if writeInterval <= 0 {
		writeInterval = defaultStatusFileWriteInterval
	}
	w.mu.Unlock()

	go w.run(stopCh, doneCh, writeInterval)
}

func (w *statusFileWriter) Stop() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	stopCh := w.stopCh
	doneCh := w.doneCh
	w.running = false
	w.stopCh = nil
	w.doneCh = nil
	w.mu.Unlock()

	close(stopCh)
	<-doneCh
	_ = os.Remove(w.path)
}

func (w *statusFileWriter) run(stopCh <-chan struct{}, doneCh chan<- struct{}, writeInterval time.Duration) {
	defer close(doneCh)

	_ = writeStatusSnapshotFile(w.path, w.snapshot())

	ticker := time.NewTicker(writeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			_ = writeStatusSnapshotFile(w.path, w.snapshot())
		}
	}
}

func writeStatusSnapshotFile(path string, snapshot StatusSnapshot) error {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".status-*.json")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}

	cleanup = false
	return nil
}
