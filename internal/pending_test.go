package internal

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// pendingFixture is a tree with one sidecar and a registry where A and B are live on it.
type pendingFixture struct {
	t          *testing.T
	dir, reg   string
	pidA, pidB int64
	sidecar    string
}

func newPendingFixture(t *testing.T) *pendingFixture {
	t.Helper()
	pf := &pendingFixture{t: t, dir: resolveDir(t.TempDir()), reg: t.TempDir(), pidA: livePID(t), pidB: livePID(t)}
	pf.sidecar = filepath.Join(pf.dir, "doc.annotations.json5")
	writeInstance(t, pf.reg, instance{PID: pf.pidA, Port: portA, Root: pf.dir})
	writeInstance(t, pf.reg, instance{PID: pf.pidB, Port: portB, Root: pf.dir})
	return pf
}

// write puts the sidecar on disk: open entries routed to A, to B, to a gone port and to none,
// then one answered and one closed, both routed to A.
func (pf *pendingFixture) write(trigger string) {
	pf.t.Helper()
	open := func(id string, via int) string {
		key := ""
		if via != 0 {
			key = fmt.Sprintf(", via: %d", via)
		}
		return fmt.Sprintf(`{id: %q, exact: null, file_hash: "", thread: [{by: "reader", at: null, text: "fix"%s}]}`, id, key)
	}
	entries := []string{
		open("toA", portA), open("toB", portB), open("gone", 7000), open("unrouted", 0),
		fmt.Sprintf(`{id: "answered", exact: null, file_hash: "", thread: [{by: "reader", at: null, text: "fix", via: %d}, {by: "agent", at: null, text: "ok"}]}`, portA),
		fmt.Sprintf(`{id: "closed", exact: null, file_hash: "", status: "done", thread: [{by: "reader", at: null, text: "fix", via: %d}]}`, portA),
	}
	body := fmt.Sprintf("// header\n{version: 2, file: \"doc.md\", trigger: %q, annotations: [%s]}", trigger, strings.Join(entries, ",\n"))
	if err := os.WriteFile(pf.sidecar, []byte(body), 0o644); err != nil {
		pf.t.Fatal(err)
	}
}

func (pf *pendingFixture) ready(content string) {
	pf.t.Helper()
	if err := os.WriteFile(filepath.Join(pf.dir, readyName), []byte(content), 0o644); err != nil {
		pf.t.Fatal(err)
	}
}

// run calls pending for port and pid on the tree, and returns the ids it listed.
func (pf *pendingFixture) run(port int, pid, claim int64) ([]string, bool, string) {
	pf.t.Helper()
	var out, errOut bytes.Buffer
	found, err := Pending{Port: port, PID: pid, Claim: claim, Registry: pf.reg, Paths: []string{pf.dir}, Out: &out, Err: &errOut}.Run()
	if err != nil {
		pf.t.Fatal(err)
	}
	var ids []string
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		if line == "" {
			continue
		}
		sc, id, ok := strings.Cut(line, "\t")
		if !ok || sc != pf.sidecar {
			pf.t.Fatalf("expected <sidecar>\\t<id>, got %q", line)
		}
		ids = append(ids, id)
	}
	if found != (len(ids) > 0) {
		pf.t.Fatalf("expected found to tell a line was printed, got %v for %v", found, ids)
	}
	return ids, found, errOut.String()
}

// 013-AC12, 013-AC6 (pending)
func TestPending(t *testing.T) {
	t.Parallel()

	pf := newPendingFixture(t)
	pf.write(triggerAuto)
	dead := int64(exitedPID(t))
	writeInstance(t, pf.reg, instance{PID: dead, Port: portC, Root: pf.dir})
	asleep := livePID(t)

	for _, tc := range []struct {
		name string
		port int
		pid  int64
		want []string
	}{
		{"A: its own, the gone and the unrouted", portA, pf.pidA, []string{"toA", "gone", "unrouted"}},
		{"B: its own only", portB, pf.pidB, []string{"toB"}},
		{"a dead PID", portC, dead, nil},
		{"a port that is not the PID's", portB, pf.pidA, nil},
		{"routing off: no file, live PID", portC, asleep, []string{"toA", "toB", "gone", "unrouted"}},
	} {
		ids, _, _ := pf.run(tc.port, tc.pid, 0)
		if !reflect.DeepEqual(ids, tc.want) {
			t.Errorf("%s: expected %v, got %v", tc.name, tc.want, ids)
		}
	}
	if _, _, stderr := pf.run(portC, dead, 0); !strings.Contains(stderr, "gone") {
		t.Errorf("expected stderr to say the instance is gone, got %q", stderr)
	}
	if _, ok := readInstanceFile(t, pf.reg, int(dead)); ok {
		t.Error("expected pending to remove the file of a dead PID")
	}

	// A stopped A: its file is there, not seen for a minute.
	writeInstance(t, pf.reg, instance{PID: pf.pidA, Port: portA, Root: pf.dir, SeenAt: time.Now().Add(-2 * registryLive).Format(time.RFC3339)})
	if ids, found, stderr := pf.run(portA, pf.pidA, 0); found || !strings.Contains(stderr, "not seen") {
		t.Errorf("expected a stopped instance to own nothing, got %v %q", ids, stderr)
	}
	// B is then the lowest live port.
	if ids, _, _ := pf.run(portB, pf.pidB, 0); !reflect.DeepEqual(ids, []string{"toA", "toB", "gone", "unrouted"}) {
		t.Errorf("expected B to own everything once A is stopped, got %v", ids)
	}
}

// 013-AC12: a sidecar that does not parse is reported and skipped, the others still listed.
func TestPendingSkipsABadSidecar(t *testing.T) {
	t.Parallel()

	pf := newPendingFixture(t)
	pf.write(triggerAuto)
	// Walked before doc.annotations.json5.
	writeFile(t, pf.dir, "aa.annotations.json5", "{ bad")
	ids, found, stderr := pf.run(portA, pf.pidA, 0)
	if !found || !reflect.DeepEqual(ids, []string{"toA", "gone", "unrouted"}) || !strings.Contains(stderr, "aa.annotations.json5") {
		t.Fatalf("expected A's entries listed and the bad sidecar on stderr, got %v %v %q", ids, found, stderr)
	}
}

// 013-AC12: manual mode lists the batch only for its owner, unclaimed.
func TestPendingManual(t *testing.T) {
	t.Parallel()

	pf := newPendingFixture(t)
	pf.write(triggerManual)
	if ids, _, _ := pf.run(portA, pf.pidA, 0); ids != nil {
		t.Fatalf("expected nothing without a ready file, got %v", ids)
	}
	pf.ready(fmt.Sprintf(`{sent_at: "2026-11-04T10:00:00Z", pid: null, via: %d}`, portA))
	if ids, _, _ := pf.run(portA, pf.pidA, 0); !reflect.DeepEqual(ids, []string{"ready", "toA", "gone", "unrouted"}) {
		t.Fatalf("expected the batch and A's entries, got %v", ids)
	}
	if ids, _, _ := pf.run(portB, pf.pidB, 0); ids != nil {
		t.Fatalf("expected nothing for B, got %v", ids)
	}
	pf.ready(fmt.Sprintf(`{sent_at: "2026-11-04T10:00:00Z", pid: %d, via: %d}`, os.Getpid(), portA))
	if ids, _, _ := pf.run(portA, pf.pidA, 0); ids != nil {
		t.Fatalf("expected nothing once claimed, got %v", ids)
	}
	// A gone via is the lowest port's.
	pf.ready(`{sent_at: "2026-11-04T10:00:00Z", pid: null, via: 7000}`)
	if ids, _, _ := pf.run(portA, pf.pidA, 0); len(ids) == 0 || ids[0] != "ready" {
		t.Fatalf("expected A to own a batch whose via is gone, got %v", ids)
	}
}

// 013-AC12
func TestPendingClaim(t *testing.T) {
	t.Parallel()

	pf := newPendingFixture(t)
	pf.write(triggerManual)
	agent := livePID(t)
	readyFile := filepath.Join(pf.dir, readyName)

	// Replaced by a batch of B between pending and the claim.
	pf.ready(fmt.Sprintf(`{sent_at: "2026-11-04T10:00:00Z", pid: null, via: %d}`, portB))
	before, _ := os.ReadFile(readyFile) //nolint:gosec // the test's own file
	var out, errOut bytes.Buffer
	found, err := Pending{Port: portA, PID: pf.pidA, Claim: agent, Registry: pf.reg, Paths: []string{pf.sidecar}, Out: &out, Err: &errOut}.Run()
	after, _ := os.ReadFile(readyFile) //nolint:gosec // idem
	if err != nil || found || out.Len() != 0 || !bytes.Equal(before, after) {
		t.Fatalf("expected the claim of B's batch refused and nothing written, got %v %v %q %q", found, err, out.String(), after)
	}

	pf.ready(fmt.Sprintf(`{sent_at: "2026-11-04T10:00:00Z", pid: null, via: %d}`, portA))
	out.Reset()
	found, err = Pending{Port: portA, PID: pf.pidA, Claim: agent, Registry: pf.reg, Paths: []string{pf.sidecar}, Out: &out, Err: &errOut}.Run()
	if err != nil || !found {
		t.Fatalf("expected the claim of A's batch to succeed, got %v %v", found, err)
	}
	if want := pf.sidecar + "\tready\n" + pf.sidecar + "\ttoA\n" + pf.sidecar + "\tgone\n" + pf.sidecar + "\tunrouted\n"; out.String() != want {
		t.Fatalf("expected the lines of pending, got %q", out.String())
	}
	rf, err := loadReady(readyFile)
	if err != nil || rf.PID == nil || *rf.PID != agent || rf.ClaimedAt == nil || rf.Via != portA || rf.SentAt == nil {
		t.Fatalf("expected pid, claimed_at written and via kept, got %+v %v", rf, err)
	}

	// Claimed already: a second claim fails.
	if found, _ := (Pending{Port: portA, PID: pf.pidA, Claim: agent, Registry: pf.reg, Paths: []string{pf.sidecar}, Out: &out, Err: &errOut}).Run(); found {
		t.Fatal("expected a second claim refused")
	}
	if _, err := (Pending{Port: portA, PID: pf.pidA, Claim: agent, Registry: pf.reg, Paths: []string{pf.dir, pf.sidecar}, Out: &out, Err: &errOut}).Run(); err == nil {
		t.Fatal("expected --claim on two sidecars refused")
	}
}

func TestFindSidecars(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "a.annotations.json5", "{}")
	writeFile(t, dir, "sub/b.annotations.json5", "{}")
	writeFile(t, dir, "sub/b.md", "# b")
	got, err := findSidecars([]string{
		dir,
		filepath.Join(dir, "sub", "b.md"),
		filepath.Join(dir, "sub", "b.annotations.ready"),
		filepath.Join(dir, "missing.annotations.json5"),
	})
	want := []string{
		filepath.Join(dir, "a.annotations.json5"),
		filepath.Join(dir, "sub", "b.annotations.json5"),
		filepath.Join(dir, "sub", "b.annotations.json5"),
		filepath.Join(dir, "sub", "b.annotations.json5"),
		filepath.Join(dir, "missing.annotations.json5"),
	}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v %v", want, got, err)
	}
}
