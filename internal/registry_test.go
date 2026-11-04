package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

const (
	portA = 6419
	portB = 6420
	portC = 6421
)

// TestSleepHelper is a live process for the tests that need a PID other than their own.
func TestSleepHelper(t *testing.T) {
	if os.Getenv("GLD_TEST_SLEEP") == "" {
		t.Skip("helper process only")
	}
	time.Sleep(time.Minute)
}

// livePID runs a process for the duration of the test and returns its PID.
func livePID(t *testing.T) int64 {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestSleepHelper$") //nolint:gosec // the test binary itself
	cmd.Env = append(os.Environ(), "GLD_TEST_SLEEP=1")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})
	return int64(cmd.Process.Pid)
}

// writeInstance writes the registry file of another instance, as it would.
func writeInstance(t *testing.T, dir string, in instance) {
	t.Helper()
	if in.SeenAt == "" {
		in.SeenAt = time.Now().Format(time.RFC3339)
	}
	if in.Pages == nil {
		in.Pages = []string{}
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, strconv.FormatInt(in.PID, 10)+".json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func readInstanceFile(t *testing.T, dir string, pid int) (instance, bool) {
	t.Helper()
	in, err := readInstance(filepath.Join(dir, strconv.Itoa(pid)+".json"))
	if os.IsNotExist(err) {
		return in, false
	}
	if err != nil {
		t.Fatal(err)
	}
	return in, true
}

func TestPeers(t *testing.T) {
	t.Parallel()

	key := filepath.FromSlash("/w/docs/specs/doc.md")
	a := instance{PID: 1, Port: portA, Root: filepath.FromSlash("/w/docs"), Pages: []string{key}}
	b := instance{PID: 2, Port: portB, Root: filepath.FromSlash("/w/docs/specs"), Pages: []string{key}}
	bClosed := instance{PID: 2, Port: portB, Root: filepath.FromSlash("/w/docs/specs")}
	c := instance{PID: 3, Port: portC, Root: filepath.FromSlash("/elsewhere"), Pages: []string{key}}
	for _, tc := range []struct {
		name           string
		self           int
		live           []instance
		via, asked     int
		owner, pick    int
		targets, live2 []int
	}{
		{"alone", portA, []instance{a}, 0, 0, portA, portA, []int{portA}, []int{portA}},
		{"shared, ask B", portA, []instance{a, b}, portB, portB, portB, portB, []int{portA, portB}, []int{portA, portB}},
		{"shared, ask a port not a target", portA, []instance{a, b, c}, portC, portC, portA, portA, []int{portA, portB}, []int{portA, portB, portC}},
		{"B without the page is no target, owns its via", portA, []instance{a, bClosed}, portB, portB, portB, portA, []int{portA}, []int{portA, portB}},
		{"gone via: lowest port", portA, []instance{a, b}, 7000, 0, portA, portA, []int{portA, portB}, []int{portA, portB}},
		{"self outside: B, open or not", portC, []instance{bClosed, c}, 0, 0, portB, portB, []int{}, []int{portB, portC}},
		{"self outside, B has it open", portC, []instance{b, c}, 0, 0, portB, portB, []int{portB}, []int{portB, portC}},
		{"no instance contains it", portC, []instance{c}, portC, portC, 0, 0, []int{}, []int{portC}},
		{"routing off", 0, nil, portA, portA, 0, 0, []int{}, []int{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := peers{self: tc.self, live: tc.live}
			if got := p.owner(key, tc.via); got != tc.owner {
				t.Errorf("owner(%d): expected %d, got %d", tc.via, tc.owner, got)
			}
			if got := p.targets(key); !reflect.DeepEqual(got, tc.targets) {
				t.Errorf("targets: expected %v, got %v", tc.targets, got)
			}
			if got := p.pick(key, tc.asked); got != tc.pick {
				t.Errorf("pick(%d): expected %d, got %d", tc.asked, tc.pick, got)
			}
			if got := p.ports(); !reflect.DeepEqual(got, tc.live2) {
				t.Errorf("ports: expected %v, got %v", tc.live2, got)
			}
		})
	}
}

// 013-AC7: a symlinked directory is followed, a symlinked file is not.
func TestPageKey(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("symlinks need privileges on Windows")
	}
	tmp := resolveDir(t.TempDir())
	writeFile(t, tmp, "real/docs/doc.md", "# doc\n")
	writeFile(t, tmp, "other/target.md", "# target\n")
	if err := os.Symlink(filepath.Join(tmp, "real"), filepath.Join(tmp, "link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(tmp, "other", "target.md"), filepath.Join(tmp, "real", "docs", "alias.md")); err != nil {
		t.Fatal(err)
	}
	direct := pageKey(http.Dir(filepath.Join(tmp, "real")), "/docs/doc.md")
	if want := filepath.Join(tmp, "real", "docs", "doc.md"); direct != want {
		t.Fatalf("expected %q, got %q", want, direct)
	}
	if got := pageKey(http.Dir(tmp), "/link/docs/doc.md"); got != direct {
		t.Errorf("through a symlinked directory: expected %q, got %q", direct, got)
	}
	alias := pageKey(http.Dir(filepath.Join(tmp, "real")), "/docs/alias.md")
	if target := pageKey(http.Dir(filepath.Join(tmp, "other")), "/target.md"); alias == target {
		t.Errorf("a symlinked file shares the key of its target: %q", alias)
	}
	if root := instanceRoot(filepath.Join(tmp, "real", "docs")); !contains(root, alias) {
		t.Errorf("expected the root %q to contain the link %q", root, alias)
	}
	if root := instanceRoot(filepath.Join(tmp, "link", "docs", "doc.md")); root != filepath.Join(tmp, "real", "docs") {
		t.Errorf("expected the root of a file argument to be its directory resolved, got %q", root)
	}
	for _, tc := range []struct {
		root, key string
		want      bool
	}{
		{"/w/docs", "/w/docs/a.md", true},
		{"/w/docs", "/w/docs/sub/a.md", true},
		{"/w/docs", "/w/a.md", false},
		{"/w/docs", "/w/docs-old/a.md", false},
		{"/", "/a.md", true},
	} {
		if got := contains(filepath.FromSlash(tc.root), filepath.FromSlash(tc.key)); got != tc.want {
			t.Errorf("contains(%q, %q): expected %v", tc.root, tc.key, tc.want)
		}
	}
}

// 013-AC4
func TestScanRegistry(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dead := exitedPID(t)
	live, stopped := livePID(t), livePID(t)
	writeInstance(t, dir, instance{PID: int64(dead), Port: portA})
	writeInstance(t, dir, instance{PID: live, Port: portC})
	writeInstance(t, dir, instance{PID: stopped, Port: portB, SeenAt: time.Now().Add(-2 * registryLive).Format(time.RFC3339)})
	// Not instance files: a temp file of replaceFile, a file of another name.
	writeFile(t, dir, fmt.Sprintf(".%d.json-123", dead), "{}")
	writeFile(t, dir, "notes.txt", "x")

	got := scanRegistry(dir)
	if len(got) != 2 || got[0].PID != stopped || got[1].PID != live {
		t.Fatalf("expected the two live PIDs by port, got %+v", got)
	}
	if _, ok := readInstanceFile(t, dir, dead); ok {
		t.Error("expected the file of a dead PID removed")
	}
	if _, ok := readInstanceFile(t, dir, int(stopped)); !ok {
		t.Error("expected the file of a stopped instance kept")
	}
	for _, name := range []string{fmt.Sprintf(".%d.json-123", dead), "notes.txt"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected %s left alone: %v", name, err)
		}
	}

	g := newRegistry(dir, portA+100, "/nowhere", &bytes.Buffer{})
	if ports := g.read().ports(); !reflect.DeepEqual(ports, []int{portC, portA + 100}) {
		t.Errorf("expected the stopped instance out of the read, got %v", ports)
	}
	if _, ok := readInstanceFile(t, dir, int(stopped)); !ok {
		t.Error("expected a read to keep the file of a stopped instance")
	}

	// A start purges too.
	writeInstance(t, dir, instance{PID: int64(dead), Port: portA})
	if err := g.start(); err != nil {
		t.Fatal(err)
	}
	if _, ok := readInstanceFile(t, dir, dead); ok {
		t.Error("expected a start to remove the file of a dead PID")
	}
}

// syncBuffer is the terminal of a served instance; started is closed at the start line.
type syncBuffer struct {
	mu      sync.Mutex
	buf     bytes.Buffer
	onStart func()
	started chan struct{}
}

func newSyncBuffer() *syncBuffer { return &syncBuffer{started: make(chan struct{})} }

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if bytes.Contains(p, []byte("Starting server")) {
		if b.onStart != nil {
			b.onStart()
		}
		close(b.started)
	}
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

var startLine = regexp.MustCompile(`🚀 Starting server: http://localhost:(\d+)/\S* \(pid (\d+)\)`)

// served is an instance running Serve in the background.
type served struct {
	s    *Server
	out  *syncBuffer
	port int
	done chan error
}

// serveInstance runs Serve on arg with the registry at reg, on a free port, and stops it at the
// end of the test; tune adjusts the server before it starts.
func serveInstance(t *testing.T, arg, reg string, reload bool, tune func(*Server)) *served {
	t.Helper()
	s := NewServer("localhost", 0, false, false, reload, NewParser(), "", reg)
	s.regEvery = 50 * time.Millisecond
	s.signals = make(chan os.Signal, 1)
	out := newSyncBuffer()
	s.out = out
	if tune != nil {
		tune(s)
	}
	inst := &served{s: s, out: out, done: make(chan error, 1)}
	go func() { inst.done <- s.Serve(arg) }()
	select {
	case <-out.started:
	case err := <-inst.done:
		t.Fatalf("Serve returned before the start line: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("no start line within 5 s")
	}
	m := startLine.FindStringSubmatch(out.String())
	if m == nil {
		t.Fatalf("expected a start line with the PID, got %q", out.String())
	}
	if m[2] != strconv.Itoa(os.Getpid()) {
		t.Fatalf("expected the PID %d in the start line, got %s", os.Getpid(), m[2])
	}
	inst.port, _ = strconv.Atoi(m[1])
	t.Cleanup(func() { inst.stop(t) })
	return inst
}

// stop sends SIGTERM and waits for Serve to return; once is enough.
func (inst *served) stop(t *testing.T) {
	t.Helper()
	select {
	case inst.s.signals <- syscall.SIGTERM:
	default:
	}
	select {
	case err, ok := <-inst.done:
		if !ok {
			return
		}
		if err != nil {
			t.Errorf("Serve: %v", err)
		}
		close(inst.done)
	case <-time.After(5 * time.Second):
		t.Error("Serve did not return within 5 s of SIGTERM")
	}
}

// 013-AC1, 013-AC5. Not parallel: every instance of the process has its PID.
func TestServeRegistersTheInstance(t *testing.T) {
	for _, reload := range []bool{true, false} {
		t.Run(fmt.Sprintf("reload %v", reload), func(t *testing.T) {
			docs := filepath.Join(t.TempDir(), "docs", "specs")
			writeFile(t, docs, "doc.md", "# doc\n")
			reg := filepath.Join(t.TempDir(), "instances")
			var atStart bool
			inst := serveInstance(t, docs, reg, reload, func(s *Server) {
				s.out.(*syncBuffer).onStart = func() {
					_, err := os.Stat(filepath.Join(reg, strconv.Itoa(os.Getpid())+".json"))
					atStart = err == nil
				}
			})
			if !atStart {
				t.Fatal("expected the registry file written before the start line")
			}
			in, ok := readInstanceFile(t, reg, os.Getpid())
			if !ok || in.Port != inst.port || in.PID != int64(os.Getpid()) || in.Root != resolveDir(docs) || in.Pages == nil || len(in.Pages) != 0 {
				t.Fatalf("expected port %d, PID, root %q and no page, got %+v", inst.port, resolveDir(docs), in)
			}
			// seen_at has a second's precision.
			first := in.SeenAt
			deadline := time.Now().Add(3 * time.Second)
			for in.SeenAt == first && time.Now().Before(deadline) {
				time.Sleep(50 * time.Millisecond)
				in, _ = readInstanceFile(t, reg, os.Getpid())
			}
			if in.SeenAt == first {
				t.Fatalf("expected seen_at rewritten, still %s", first)
			}

			inst.stop(t)
			if _, ok := readInstanceFile(t, reg, os.Getpid()); ok {
				t.Fatal("expected the registry file removed on SIGTERM")
			}
		})
	}
}

// 013-AC2
func TestServeOnATakenPort(t *testing.T) {
	docs := t.TempDir()
	writeFile(t, docs, "doc.md", "# doc\n")
	reg := filepath.Join(t.TempDir(), "instances")
	first := serveInstance(t, docs, reg, false, nil)
	first.stop(t)

	taken := serveInstance(t, docs, t.TempDir(), false, nil)
	s := NewServer("localhost", taken.port, false, false, false, NewParser(), "", reg)
	out := newSyncBuffer()
	s.out = out
	if err := s.Serve(docs); err == nil {
		t.Fatal("expected Serve to fail on a taken port")
	}
	if out.String() != "" {
		t.Errorf("expected no start line, got %q", out.String())
	}
	if _, ok := readInstanceFile(t, reg, os.Getpid()); ok {
		t.Error("expected no registry file")
	}
}

// dialInstance opens the reload socket of page on a served instance.
func dialInstance(t *testing.T, port int, page string) *websocket.Conn {
	t.Helper()
	conn, resp, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://127.0.0.1:%d%s?page=%s", port, reloadEndpoint, page), nil)
	if resp != nil {
		resp.Body.Close() //nolint:errcheck
	}
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

// 013-AC3
func TestRegistryListsTheOpenPages(t *testing.T) {
	docs := t.TempDir()
	writeFile(t, docs, "doc.md", "# doc\n")
	reg := t.TempDir()
	inst := serveInstance(t, filepath.Join(docs, "doc.md"), reg, true, func(s *Server) {
		s.regEvery = time.Hour // only the page changes write
		s.reload.releaseAfter = 500 * time.Millisecond
	})
	key := filepath.Join(resolveDir(docs), "doc.md")
	listed := func() bool {
		in, _ := readInstanceFile(t, reg, os.Getpid())
		return reflect.DeepEqual(in.Pages, []string{key})
	}

	conn := dialInstance(t, inst.port, "/doc.md")
	eventually(t, "the page listed at once", listed)
	// A reload: the socket closes, another opens within the grace.
	conn.Close() //nolint:errcheck
	time.Sleep(200 * time.Millisecond)
	conn = dialInstance(t, inst.port, "/doc.md")
	time.Sleep(500 * time.Millisecond)
	if !listed() {
		t.Fatal("expected a reload to keep the page listed")
	}
	conn.Close() //nolint:errcheck
	eventually(t, "the page dropped after the grace", func() bool { return !listed() })
}

// 013-AC6 (server part)
func TestReadOnlyRegistry(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs POSIX permissions enforced on a non-root user")
	}
	logs := captureLog(t)
	ro := filepath.Join(t.TempDir(), "ro")
	if err := os.Mkdir(ro, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(ro, 0o700) }) //nolint:gosec // let TempDir clean up
	docs := t.TempDir()
	writeFile(t, docs, "doc.md", "# doc\n\nalpha\n")
	inst := serveInstance(t, filepath.Join(docs, "doc.md"), filepath.Join(ro, "instances"), true, nil)
	conn := dialInstance(t, inst.port, "/doc.md")
	defer conn.Close() //nolint:errcheck
	time.Sleep(200 * time.Millisecond)

	base := fmt.Sprintf("http://127.0.0.1:%d", inst.port)
	resp, err := http.Get(base + "/doc.md")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected the page served, got %v %v", resp, err)
	}
	resp.Body.Close() //nolint:errcheck
	resp, err = http.Get(base + "/__annotations?path=/doc.md")
	if err != nil {
		t.Fatal(err)
	}
	var got annotationsResponse
	err = json.NewDecoder(resp.Body).Decode(&got)
	resp.Body.Close() //nolint:errcheck
	if err != nil || got.Self != inst.port || len(got.Targets) != 0 || len(got.Live) != 0 {
		t.Fatalf("expected self and no target, got %+v %v", got.routing, err)
	}
	if n := strings.Count(logs.String(), "instance registry off"); n != 1 {
		t.Fatalf("expected one log line, got %d: %s", n, logs)
	}
}

// routedFixture is a diffFixture whose server is instance A of a registry of its own, with the
// pages the test sets.
type routedFixture struct {
	*diffFixture
	reg   string
	g     *registry
	key   string
	out   *bytes.Buffer
	mu    sync.Mutex
	open  map[string]string
	notes []string
}

func newRoutedFixture(t *testing.T, content string) *routedFixture {
	t.Helper()
	f := &diffFixture{t: t, dir: t.TempDir()}
	s := NewServer("localhost", portA, false, false, false, NewParser(), "", "")
	f.handler = s.newHandler(http.Dir(f.dir))
	f.write(content)
	rf := &routedFixture{diffFixture: f, reg: t.TempDir(), key: pageKey(http.Dir(f.dir), docPath), out: &bytes.Buffer{}, open: map[string]string{}}
	rf.g = newRegistry(rf.reg, portA, resolveDir(f.dir), rf.out)
	rf.g.pages = func() map[string]string {
		rf.mu.Lock()
		defer rf.mu.Unlock()
		out := map[string]string{}
		for k, v := range rf.open {
			out[k] = v
		}
		return out
	}
	rf.g.notify = func(page string) {
		rf.mu.Lock()
		defer rf.mu.Unlock()
		rf.notes = append(rf.notes, page)
	}
	if err := rf.g.start(); err != nil {
		t.Fatal(err)
	}
	s.reg = rf.g
	return rf
}

func (rf *routedFixture) openPage() {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	rf.open[docPath] = rf.key
}

func (rf *routedFixture) notified() int {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return len(rf.notes)
}

// peer writes the file of another live instance listing the page, or not.
func (rf *routedFixture) peer(pid int64, port int, root string, listed bool) {
	rf.t.Helper()
	in := instance{PID: pid, Port: port, Root: root}
	if listed {
		in.Pages = []string{rf.key}
	}
	writeInstance(rf.t, rf.reg, in)
}

// 013-AC7
func TestSharedPage(t *testing.T) {
	t.Parallel()

	rf := newRoutedFixture(t, "# Title\n")
	// Listed or not (--no-reload), A is the first of its own targets.
	if got := rf.getAnnotations(); got.Self != portA || !reflect.DeepEqual(got.Targets, []int{portA}) {
		t.Fatalf("expected A alone, got %+v", got.routing)
	}
	pidB, pidC := livePID(t), livePID(t)
	rf.peer(pidB, portB, filepath.Dir(rf.key), true)
	// C has the page open, but its agent does not watch the file.
	rf.peer(pidC, portC, filepath.Join(filepath.Dir(rf.key), "sub"), true)
	got := rf.getAnnotations()
	if got.Self != portA || !reflect.DeepEqual(got.Targets, []int{portA, portB}) || !reflect.DeepEqual(got.Live, []int{portA, portB, portC}) {
		t.Fatalf("expected targets [A B] and C live, got %+v", got.routing)
	}
	rec := rf.ready(http.MethodDelete)
	var ready readyResponse
	if json.Unmarshal(rec.Body.Bytes(), &ready) != nil || !reflect.DeepEqual(ready.Targets, []int{portA, portB}) {
		t.Fatalf("expected the ready answer to carry the targets, got %s", rec.Body)
	}

	rf.openPage()
	for range 3 {
		rf.g.tick()
	}
	warnB := fmt.Sprintf("⚠ doc.md also open in :%d (pid %d)\n", portB, pidB)
	warnC := fmt.Sprintf("⚠ doc.md also open in :%d (pid %d)\n", portC, pidC)
	if out := rf.out.String(); strings.Count(out, warnB) != 1 || strings.Count(out, warnC) != 1 {
		t.Fatalf("expected one warning per other instance, got %q", out)
	}
	if in, _ := readInstanceFile(t, rf.reg, os.Getpid()); !reflect.DeepEqual(in.Pages, []string{rf.key}) {
		t.Fatalf("expected A to list the page, got %+v", in)
	}
}

// The sockets of a page are told when its targets change at a registry read, not otherwise.
func TestTargetsChangeNotifiesThePage(t *testing.T) {
	t.Parallel()

	rf := newRoutedFixture(t, "# Title\n")
	rf.openPage()
	rf.g.tick()
	rf.g.tick()
	if n := rf.notified(); n != 0 {
		t.Fatalf("expected no notice while A is alone, got %d", n)
	}
	pidB := livePID(t)
	rf.peer(pidB, portB, filepath.Dir(rf.key), true)
	rf.g.tick()
	rf.g.tick()
	if n := rf.notified(); n != 1 {
		t.Fatalf("expected one notice when B opens the page, got %d", n)
	}
	if err := os.Remove(filepath.Join(rf.reg, strconv.FormatInt(pidB, 10)+".json")); err != nil {
		t.Fatal(err)
	}
	rf.g.tick()
	if n := rf.notified(); n != 2 {
		t.Fatalf("expected a notice when B goes, got %d", n)
	}
}
