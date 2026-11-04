package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/titanous/json5"
)

const (
	readySuffix = ".annotations.ready"
	// Two fields and the agent's claimed_at; anything bigger is not ours or the agent's.
	maxReadyBytes = 4 << 10

	triggerAuto   = "auto"
	triggerManual = "manual"

	readyNone    = "none"
	readySent    = "sent"
	readyWorking = "working"
	readyGone    = "gone"
)

var (
	errNotManual    = errors.New("trigger is not manual")
	errNothingOpen  = errors.New("no open entry")
	errNotOwned     = errors.New("nothing open for")
	errBatchPending = errors.New("batch pending for")
)

// readyFile is what the preview writes when the reader sends a batch; the agent sets PID
// and claimed_at when it takes it. Via is the port of the instance the batch is for.
type readyFile struct {
	SentAt *string `json:"sent_at"`
	// int64, not int: a 32-bit build would refuse a large PID and read the file as sent.
	PID       *int64  `json:"pid"`
	ClaimedAt *string `json:"claimed_at,omitempty"`
	Via       int     `json:"via,omitempty"`
}

// readyState is the ready file as the page shows it.
type readyState struct {
	State  string  `json:"state"`
	PID    *int64  `json:"pid"`
	SentAt *string `json:"sent_at"`
	Via    int     `json:"via,omitempty"`
}

// readyResponse is the ready state with the registry read, as the annotations GET has it.
type readyResponse struct {
	readyState
	routing
}

// readyPath is the ready file beside target's sidecar, confined like it.
func readyPath(dir http.Dir, target string) string {
	return strings.TrimSuffix(sidecarPath(dir, target), sidecarSuffix) + readySuffix
}

// readReady reports the state of the ready file at p. A file that does not parse reads as
// sent: the agent writes it too, a bad write of its own must not break the page.
func readReady(p string) readyState {
	switch regular, err := isRegular(p); {
	case errors.Is(err, os.ErrNotExist):
		return readyState{State: readyNone}
	case err != nil, !regular:
		return readyState{State: readySent}
	}
	rf, err := loadReady(p)
	if err != nil {
		return readyState{State: readySent}
	}
	st := readyState{PID: rf.PID, SentAt: rf.SentAt, Via: rf.Via}
	switch {
	case rf.PID == nil:
		st.State = readySent
	case pidAlive(*rf.PID):
		st.State = readyWorking
	default:
		st.State = readyGone
	}
	return st
}

// loadReady parses the ready file at p; check isRegular first.
func loadReady(p string) (readyFile, error) {
	var rf readyFile
	f, err := os.Open(p) //nolint:gosec // p is confined to dir by readyPath, or named by the agent
	if err != nil {
		return rf, err
	}
	data, err := io.ReadAll(io.LimitReader(f, maxReadyBytes))
	_ = f.Close()
	if err == nil {
		err = json5.Unmarshal(data, &rf)
	}
	return rf, err
}

// writeReady replaces the ready file at p whole: the agent may stat it at any time.
func writeReady(p string, rf readyFile) error {
	data, err := json.Marshal(rf)
	if err != nil {
		return err
	}
	if err := replaceFile(p, "."+filepath.Base(p)+"-*", 0o644, append(data, '\n')); err != nil {
		return fmt.Errorf("write %s: %w", p, err)
	}
	return nil
}

// lastVia is where the last message of an open entry is routed.
func lastVia(a annotation) int {
	return a.Thread[len(a.Thread)-1].Via
}

// isRegular is checked before any open: opening a FIFO blocks until a writer comes, and a
// send would hang holding the sidecar lock. ponytail: Lstat then Open, a swap in between wins.
func isRegular(p string) (bool, error) {
	fi, err := os.Lstat(p)
	return err == nil && fi.Mode().IsRegular(), err
}

// isOpen tells a request from a handled entry: the reader spoke last and the agent did not
// close it.
func isOpen(a annotation) bool {
	return a.Status != "done" && len(a.Thread) > 0 && a.Thread[len(a.Thread)-1].By == byReader
}

// sendReady writes the ready file of target for the instance rt.via, under the sidecar's lock,
// the one the agent claims under. A batch already sent or worked on is left as is when its
// owner is rt.via, its gone via rewritten; refused when it is another instance's.
func sendReady(dir http.Dir, target string, rt route) (readyState, error) {
	// O_RDWR: on NFS, flock is emulated by fcntl, whose exclusive lock needs a writable fd.
	f, err := lockSidecar(sidecarPath(dir, target), os.O_RDWR, true)
	if errors.Is(err, os.ErrNotExist) {
		return readyState{}, errNotManual
	}
	if err != nil {
		return readyState{}, err
	}
	defer releaseSidecar(f)

	sc, err := decodeSidecar(f, target)
	if err != nil {
		return readyState{}, err
	}
	if sc.Trigger != triggerManual {
		return readyState{}, errNotManual
	}
	open, owned := false, false
	for _, a := range sc.Annotations {
		if isOpen(a) {
			open = true
			owned = owned || rt.owner == nil || rt.owner(lastVia(a)) == rt.via
		}
	}
	if !open {
		return readyState{}, errNothingOpen
	}

	rp := readyPath(dir, target)
	// Not a regular file is not a batch of ours: replaced, the rename does not follow it.
	if regular, _ := isRegular(rp); regular {
		if st := readReady(rp); st.State == readySent || st.State == readyWorking {
			if rt.owner == nil {
				return st, nil
			}
			if o := rt.owner(st.Via); o != rt.via {
				return readyState{}, fmt.Errorf("%w :%d", errBatchPending, o)
			}
			// A garbage file has no via to rewrite: left to the agent that wrote it.
			rf, err := loadReady(rp)
			if st.Via == rt.via || err != nil {
				return st, nil
			}
			rf.Via = rt.via
			if err := writeReady(rp, rf); err != nil {
				return readyState{}, err
			}
			st.Via = rt.via
			return st, nil
		}
	}
	if !owned {
		return readyState{}, fmt.Errorf("%w :%d", errNotOwned, rt.via)
	}
	now := time.Now().Format(time.RFC3339)
	if err := writeReady(rp, readyFile{SentAt: &now, Via: rt.via}); err != nil {
		return readyState{}, err
	}
	return readyState{State: readySent, SentAt: &now, Via: rt.via}, nil
}

// cancelReady removes the ready file under the sidecar's lock: unlocked, an agent claiming
// at that moment would write the cancelled batch back.
func cancelReady(dir http.Dir, target string) error {
	f, err := lockSidecar(sidecarPath(dir, target), os.O_RDWR, true)
	switch {
	case err == nil:
		defer releaseSidecar(f)
	case !errors.Is(err, os.ErrNotExist):
		return err
	}
	return removeReady(readyPath(dir, target))
}

// removeReady deletes the ready file; already gone is fine.
func removeReady(p string) error {
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// handleReady sends (POST) or cancels (DELETE) the batch of a markdown file in manual mode.
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request, dir http.Dir, regex *regexp.Regexp) {
	if !sameOrigin(r) {
		http.Error(w, "cross-origin request refused", http.StatusForbidden)
		return
	}
	target := r.URL.Query().Get("path")
	if !isMarkdownTarget(target, regex) || isRevisionTarget(target) {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	if _, err := readToString(dir, target); err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	// A POST body names the instance the reader chose; none is the first target.
	var body struct {
		Via int `json:"via"`
	}
	if r.Method == http.MethodPost && r.ContentLength != 0 {
		if err := decodeJSONBody(w, r, maxReadyBytes, &body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
	}
	rt, info := s.route(dir, target, body.Via)
	var st readyState
	var err error
	switch r.Method {
	case http.MethodPost:
		st, err = sendReady(dir, target, rt)
	case http.MethodDelete:
		st = readyState{State: readyNone}
		err = cancelReady(dir, target)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	switch {
	case errors.Is(err, errNotManual), errors.Is(err, errNothingOpen), errors.Is(err, errNotOwned), errors.Is(err, errBatchPending):
		http.Error(w, err.Error(), http.StatusConflict)
		return
	case err != nil:
		log.Println(err)
		http.Error(w, "ready file not written", http.StatusInternalServerError)
		return
	}

	setNoCacheHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	resp := readyResponse{readyState: st, routing: info}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Println(err)
	}
}
