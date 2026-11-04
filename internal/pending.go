package internal

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/titanous/json5"
)

// Pending is one run of `grip-live-diff pending`: the open entries of Paths whose owner is the
// instance Port / PID, one "<sidecar>\t<entry id>" line each on Out, and in manual mode the
// ready file it may claim. Claim, the PID of the agent, takes that ready file.
type Pending struct {
	Port     int
	PID      int64
	Claim    int64
	Registry string
	Paths    []string
	Out, Err io.Writer
}

// Run prints the lines and tells whether it printed one. A sidecar it cannot read is reported on
// Err and skipped.
func (p Pending) Run() (bool, error) {
	if !pidAlive(p.PID) {
		_, _ = fmt.Fprintf(p.Err, "instance :%d (pid %d) is gone: nothing is its own\n", p.Port, p.PID)
		// Its file, if any, goes with the other dead ones.
		scanRegistry(p.Registry)
		return false, nil
	}
	owns := p.owns()
	sidecars, err := findSidecars(p.Paths)
	if err != nil {
		return false, err
	}
	if p.Claim != 0 && len(sidecars) != 1 {
		return false, fmt.Errorf("--claim takes one sidecar, got %d", len(sidecars))
	}
	found := false
	for _, sc := range sidecars {
		lines, err := pendingLines(sc, owns, p.Claim)
		if err != nil {
			// One bad sidecar must not hide the entries of the others.
			_, _ = fmt.Fprintln(p.Err, err)
			continue
		}
		for _, l := range lines {
			if _, err := fmt.Fprintf(p.Out, "%s\t%s\n", sc, l); err != nil {
				return found, err
			}
			found = true
		}
	}
	return found, nil
}

// owns applies the owner rule for the instance, once the registry is purged. With no file for
// it, the instance runs with routing off and owns everything; a file not seen for registryLive
// is a stopped instance and owns nothing.
func (p Pending) owns() func(key string, via int) bool {
	all := scanRegistry(p.Registry)
	now := time.Now()
	i := slices.IndexFunc(all, func(in instance) bool { return in.PID == p.PID })
	if i < 0 {
		return func(string, int) bool { return true }
	}
	switch in := all[i]; {
	case in.Port != p.Port:
		_, _ = fmt.Fprintf(p.Err, "pid %d serves :%d, not :%d: nothing is its own\n", p.PID, in.Port, p.Port)
		return func(string, int) bool { return false }
	case !fresh(in, now, registryLive):
		_, _ = fmt.Fprintf(p.Err, "instance :%d (pid %d) not seen since %s: nothing is its own\n", p.Port, p.PID, in.SeenAt)
		return func(string, int) bool { return false }
	}
	v := peers{live: liveOf(all, now, registryLive)}
	return func(key string, via int) bool { return v.owner(key, via) == p.Port }
}

// findSidecars maps each path to sidecars: a directory to the ones under it, a markdown or
// ready file to the sidecar beside it.
func findSidecars(paths []string) ([]string, error) {
	var out []string
	for _, p := range paths {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			// The trailing separator makes WalkDir follow a symlinked directory: it Lstats its root.
			err := filepath.WalkDir(p+string(filepath.Separator), func(q string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && strings.HasSuffix(q, sidecarSuffix) {
					out = append(out, q)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
			continue
		}
		switch {
		case strings.HasSuffix(p, sidecarSuffix):
			out = append(out, p)
		case strings.HasSuffix(p, readySuffix):
			out = append(out, sidecarOfReady(p))
		default:
			out = append(out, sidecarOf(p))
		}
	}
	return out, nil
}

// pendingLines is what the instance owns in the sidecar at p: its open entries in auto; in
// manual, "ready" and those entries when the ready file is unclaimed and its own, nothing
// otherwise. With claim, it writes claim as the PID of the ready file first, under the lock.
func pendingLines(p string, owns func(key string, via int) bool, claim int64) ([]string, error) {
	flag, exclusive := os.O_RDONLY, false
	if claim != 0 {
		flag, exclusive = os.O_RDWR, true
	}
	f, err := lockSidecar(p, flag, exclusive)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer releaseSidecar(f)
	data, err := io.ReadAll(io.LimitReader(f, maxSidecarBytes))
	if err != nil {
		return nil, err
	}
	// The key is the one of the page: the file name as served, which the name of the sidecar
	// has lost the extension of.
	var head struct {
		File string `json:"file"`
	}
	name := strings.TrimSuffix(filepath.Base(p), sidecarSuffix) + ".md"
	if json5.Unmarshal(data, &head) == nil && head.File != "" {
		name = filepath.Base(head.File)
	}
	sc, err := parseSidecar(data, "/"+name, p)
	if err != nil {
		return nil, err
	}
	key := pageKey(http.Dir(filepath.Dir(p)), name)

	var lines []string
	for _, a := range sc.Annotations {
		if isOpen(a) && owns(key, lastVia(a)) {
			lines = append(lines, a.ID)
		}
	}
	if sc.Trigger != triggerManual {
		if claim != 0 {
			return nil, nil
		}
		return lines, nil
	}
	rp := readyOf(p)
	if regular, _ := isRegular(rp); !regular {
		return nil, nil
	}
	rf, err := loadReady(rp)
	if err != nil || rf.PID != nil || !owns(key, rf.Via) {
		return nil, nil //nolint:nilerr // a ready file that does not parse is no batch to take
	}
	if claim != 0 {
		now := time.Now().Format(time.RFC3339)
		rf.PID, rf.ClaimedAt = &claim, &now
		if err := writeReady(rp, rf); err != nil {
			return nil, err
		}
	}
	return append([]string{"ready"}, lines...), nil
}
