package file

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var ErrNoSavedOffset = fmt.Errorf("no saved offset")

var knownInfraContainers = []string{
	"istio-proxy",
	"coredns",
	"etcd",
	"kube-proxy",
	"metrics-server",
	"loki",
	"grafana",
	"flox",
}

type Tailer struct {
	path        string
	namespace   string
	trackOffset bool
	startFrom   string
	files       map[string]*os.File
	lock        sync.Mutex
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewTailer(path string, namespace string, track bool, startFrom string) *Tailer {
	ctx, cancel := context.WithCancel(context.Background())
	return &Tailer{
		path:        path,
		namespace:   namespace,
		trackOffset: track,
		startFrom:   startFrom,
		files:       make(map[string]*os.File),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func stripKubernetesPrefix(line []byte) []byte {
	idx := bytes.IndexByte(line, '{')
	if idx != -1 {
		return line[idx:]
	}
	return line
}

func isRelevantLog(filePath string, allowedNamespace string) bool {
	dir := filepath.Dir(filePath)
	podDir := filepath.Dir(dir)
	podNameFull := filepath.Base(podDir)

	parts := strings.SplitN(podNameFull, "_", 3)
	if len(parts) < 3 {
		return false
	}

	namespace, podName := parts[0], parts[1]
	if allowedNamespace != "" && namespace != allowedNamespace {
		return false
	}

	for _, infra := range knownInfraContainers {
		if strings.HasPrefix(podName, infra) {
			return false
		}
	}

	return true
}

func (t *Tailer) openFile(filePath string, handler HandlerFunc) {
	log.Printf("[Tailing] Attempting to open: %s", filePath)

	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("[Tailing] Failed to open file %s: %v", filePath, err)
		return
	}

	t.registerFile(filePath, f)
	defer t.unregisterFile(filePath)

	t.handleSeek(filePath, f)

	log.Printf("[Tailing] Start tailing: %s", filePath)

	reader := bufio.NewReader(f)
	var warned bool

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(1 * time.Second)

				if t.isFileRotated(filePath, f) {
					log.Printf("[Tailing] File rotated: reopening %s", filePath)
					t.unregisterFile(filePath)
					go t.openFile(filePath, handler)
					return
				}
				reader = bufio.NewReader(f)
				continue
			}

			log.Printf("[Taililng] Error reading file %s: %v", filePath, err)
			break
		}

		cleanLine := stripKubernetesPrefix(line)

		var event map[string]interface{}
		if err := json.Unmarshal(cleanLine, &event); err != nil {
			if !warned {
				log.Printf("[Tailing] Skipping invalid JSON in %s: %s", filePath, string(line))
				warned = true
			}
			continue
		}

		warned = false
		handler(event)

		if t.trackOffset {
			t.saveCurrentOffset(filePath, f)
		}
	}
}

func (t *Tailer) registerFile(filePath string, f *os.File) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.files[filePath] = f
}

func (t *Tailer) unregisterFile(filePath string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if f, ok := t.files[filePath]; ok && f != nil {
		defer func() {
			if err := f.Close(); err != nil {
				log.Printf("[Tailing] Failed to close file %s: %v", filePath, err)
			}
		}()
	}
	delete(t.files, filePath)
}

func (t *Tailer) seekFromSavedOffset(filePath string, f *os.File) error {
	state := loadState()
	if offset, ok := state[filePath]; ok {
		log.Printf("[Tailing] Resuming %s from offset: %d", filePath, offset)
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			log.Printf("[Tailing] Seek failed for %s: %v", filePath, err)
			return err
		}
		return nil
	}
	// No saved offset
	return ErrNoSavedOffset // fmt.Errorf("no saved offset for %s", filePath)
}

func (t *Tailer) saveCurrentOffset(filePath string, f *os.File) {
	offset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		log.Printf("[Tailing] Failed to get current offset for %s: %v", filePath, err)
		return
	}
	saveOffset(filePath, offset)
}

func (t *Tailer) isFileRotated(filePath string, f *os.File) bool {
	stat1, err1 := f.Stat()
	stat2, err2 := os.Stat(filePath)
	if err1 != nil || err2 != nil {
		return false
	}
	return !os.SameFile(stat1, stat2)
}

func (t *Tailer) scanForNewFiles(handler HandlerFunc) {
	// log.Printf("[Tailing] Watching for new files matching: %s", t.path)

	matches, err := filepath.Glob(t.path)
	if err != nil {
		log.Printf("Glob error: %v", err)
		// time.Sleep(5 * time.Second)
		return
	}

	t.lock.Lock()
	defer t.lock.Unlock()

	if len(matches) > len(t.files) {
		log.Printf("[Tailing] New files detected: %d files now being tailed", len(matches))
	}

	for _, filePath := range matches {
		if _, alreadyTailing := t.files[filePath]; alreadyTailing {
			// log.Printf("[Tailing] Already tailing: %s", filePath)
			continue
		}

		if !isRelevantLog(filePath, t.namespace) {
			log.Printf("[Tailling] Skipping irrelevant log file: %s", filePath)
			t.files[filePath] = nil // mark as ignored
			continue
		}
		log.Printf("[Tailing] New file detected: %s", filePath)
		// t.files[filePath] = nil
		go t.openFile(filePath, handler)
	}
	// t.lock.Unlock()
}

func (t *Tailer) handleSeek(filePath string, f *os.File) {
	if !t.trackOffset {
		return
	}

	effectiveStartFrom := t.startFrom
	if effectiveStartFrom == "" {
		effectiveStartFrom = "latest"
	}

	err := t.seekFromSavedOffset(filePath, f)
	if err == nil {
		log.Printf("[Tailing] Resumed from saved offset for: %s", filePath)
		return
	}

	if errors.Is(err, ErrNoSavedOffset) {
		switch effectiveStartFrom {
		case "latest":
			if _, err := f.Seek(0, io.SeekEnd); err != nil {
				log.Printf("[Tailing] Failed to seek to end for %s: %v", filePath, err)
			} else {
				log.Printf("[Tailing] No saved offset for %s, starting from end of file", filePath)
			}
		default:
			log.Printf("[Tailing] No offset found. Starting from beginning: %s", filePath)
		}
	} else {
		log.Printf("[Tailing] Failed to resume from saved offset: %v", err)
	}
}

func (t *Tailer) Run(ctx context.Context, handler HandlerFunc) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-t.ctx.Done():
			log.Println("[Tailer] Shutdown requested. Exiting Run loop.")
			t.Shutdown()
			return

		case <-ticker.C:
			t.scanForNewFiles(handler)
		}

		// time.Sleep(10 * time.Second)
	}
}

func (t *Tailer) Shutdown() {
	t.cancel()
}
