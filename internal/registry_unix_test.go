//go:build unix

package internal

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// An entry that is not a regular file is skipped: a FIFO would block every scan, a directory
// of a dead PID cannot be removed and would log on each one. Not parallel: it captures the log.
func TestScanRegistrySkipsNotARegularFile(t *testing.T) {
	var logged bytes.Buffer
	log.SetOutput(&logged)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	dir := t.TempDir()
	live := livePID(t)
	if err := syscall.Mkfifo(filepath.Join(dir, fmt.Sprintf("%d.json", live)), 0o644); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, fmt.Sprintf("%d.json/x", exitedPID(t)), "x")

	got := make(chan []instance, 1)
	go func() { got <- scanRegistry(dir) }()
	select {
	case in := <-got:
		if len(in) != 0 {
			t.Fatalf("expected no instance, got %+v", in)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected the scan not to block on a FIFO")
	}
	scanRegistry(dir)
	if logged.Len() != 0 {
		t.Fatalf("expected nothing logged, got %q", logged.String())
	}
}
