package retroarch

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"kyberix.com/gemu/internal/profiles"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func zipBytes(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	z := zip.NewWriter(&buffer)
	for name, text := range files {
		f, err := z.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Store})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = f.Write([]byte(text)); err != nil {
			t.Fatal(err)
		}
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}
func packageTransport(t *testing.T, failCore bool) transport {
	core := zipBytes(t, map[string]string{"snes9x_libretro.dll": "core fixture"})
	return func(request *http.Request) (*http.Response, error) {
		path := request.URL.Path
		data := []byte("archive fixture")
		code := 200
		switch {
		case path == "/stable/":
			data = []byte(`<a href="1.9.10/">old</a><a href="/stable/1.22.2/">stable</a><a href="1.9.9/">old</a>`)
		case strings.HasSuffix(path, ".dll.zip"):
			data = core
			if failCore && request.Method == "GET" {
				code = 503
			}
		case strings.HasSuffix(path, ".info"):
			data = []byte(`display_name = "Snes9x"`)
		}
		return &http.Response{StatusCode: code, ContentLength: int64(len(data)), Body: io.NopCloser(bytes.NewReader(data)), Header: make(http.Header), Request: request}, nil
	}
}
func waitJob(t *testing.T, s *Service, id, key string) Job {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		job, err := s.Job(id, key)
		if err != nil {
			t.Fatal(err)
		}
		if terminal(job.State) {
			return job
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("installation did not finish")
	return Job{}
}
func fakeExtractor(t *testing.T) func(context.Context, string, string, func(string)) error {
	t.Helper()
	return func(ctx context.Context, _ string, destination string, progress func(string)) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		writeFile(t, filepath.Join(destination, "RetroArch-Win64", "retroarch.exe"), []byte("official exe fixture"))
		writeFile(t, filepath.Join(destination, "RetroArch-Win64", "COPYING"), []byte("upstream license fixture"))
		progress("retroarch.exe")
		return nil
	}
}
func TestOfficialPlanInstallAndCoresOnlyOfflineReuse(t *testing.T) {
	s, id, config, _ := fixture(t)
	s.client = &http.Client{Transport: packageTransport(t, false)}
	s.extract = fakeExtractor(t)
	plan, err := s.Plan(context.Background(), id, InstallRequest{InstallApp: true, Systems: []string{"snes", "snes"}})
	if err != nil || plan.Version != "1.22.2" || len(plan.Downloads) != 3 {
		t.Fatal(plan, err)
	}
	if _, err = s.Start(context.Background(), "another-player", plan.Token); err == nil {
		t.Fatal("another player used plan")
	}
	job, err := s.Start(context.Background(), id, plan.Token)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Start(context.Background(), id, plan.Token); err == nil {
		t.Fatal("reused installation token")
	}
	result := waitJob(t, s, id, job.ID)
	if result.State != "ready" || !within(s.managedDir(), result.Executable) {
		t.Fatal(result)
	}
	if _, err = os.Stat(filepath.Join(filepath.Dir(filepath.Dir(result.Executable)), "gemu-install.json")); err != nil {
		t.Fatal("missing download receipt", err)
	}
	if _, err = os.Stat(filepath.Join(s.managedDir(), "cores", "snes9x_libretro.dll")); err != nil {
		t.Fatal(err)
	}
	s.client = &http.Client{Transport: transport(func(*http.Request) (*http.Response, error) {
		t.Error("offline reuse accessed network")
		return nil, errors.New("offline")
	})}
	plan, err = s.Plan(context.Background(), id, InstallRequest{Executable: result.Executable, Systems: []string{"snes"}})
	if err != nil || len(plan.Downloads) != 0 {
		t.Fatal(plan, err)
	}
	job, err = s.Start(context.Background(), id, plan.Token)
	if err != nil {
		t.Fatal(err)
	}
	result = waitJob(t, s, id, job.ID)
	if result.State != "ready" {
		t.Fatal(result)
	}
	data, _ := os.ReadFile(config.Executable)
	if string(data) != "existing executable" {
		t.Fatal("changed external executable")
	}
}
func TestInstallFailurePreservesExistingFilesAndCanRetry(t *testing.T) {
	s, id, config, _ := fixture(t)
	s.client = &http.Client{Transport: packageTransport(t, true)}
	s.extract = fakeExtractor(t)
	request := InstallRequest{InstallApp: true, Systems: []string{"snes"}}
	plan, err := s.Plan(context.Background(), id, request)
	if err != nil {
		t.Fatal(err)
	}
	job, err := s.Start(context.Background(), id, plan.Token)
	if err != nil {
		t.Fatal(err)
	}
	result := waitJob(t, s, id, job.ID)
	if result.State != "error" {
		t.Fatal(result)
	}
	if _, err = os.Stat(filepath.Join(s.managedDir(), "releases", "1.22.2")); !os.IsNotExist(err) {
		t.Fatal("published incomplete release")
	}
	data, _ := os.ReadFile(config.Executable)
	if string(data) != "existing executable" {
		t.Fatal("changed external install")
	}
	s.client = &http.Client{Transport: packageTransport(t, false)}
	plan, err = s.Plan(context.Background(), id, request)
	if err != nil {
		t.Fatal(err)
	}
	job, err = s.Start(context.Background(), id, plan.Token)
	if err != nil {
		t.Fatal(err)
	}
	if result = waitJob(t, s, id, job.ID); result.State != "ready" {
		t.Fatal(result)
	}
	entries, err := os.ReadDir(s.managedDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".stage-") {
			t.Fatal("staging directory survived terminal job")
		}
	}
}
func TestInstallationCancellationReleasesLock(t *testing.T) {
	s, id, _, _ := fixture(t)
	started := make(chan struct{})
	normal := packageTransport(t, false)
	s.client = &http.Client{Transport: transport(func(request *http.Request) (*http.Response, error) {
		if request.Method == "GET" && strings.HasSuffix(request.URL.Path, ".7z") {
			close(started)
			<-request.Context().Done()
			return nil, request.Context().Err()
		}
		return normal(request)
	})}
	plan, err := s.Plan(context.Background(), id, InstallRequest{InstallApp: true, Systems: []string{"snes"}})
	if err != nil {
		t.Fatal(err)
	}
	job, err := s.Start(context.Background(), id, plan.Token)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("download did not start")
	}
	if err = s.Cancel("other-player", job.ID); err == nil {
		t.Fatal("another player cancelled installation")
	}
	if err = s.Cancel(id, job.ID); err != nil {
		t.Fatal(err)
	}
	if result := waitJob(t, s, id, job.ID); result.State != "cancelled" {
		t.Fatal(result)
	}
	s.client = &http.Client{Transport: normal}
	s.extract = fakeExtractor(t)
	plan, err = s.Plan(context.Background(), id, InstallRequest{InstallApp: true, Systems: []string{"snes"}})
	if err != nil {
		t.Fatal(err)
	}
	job, err = s.Start(context.Background(), id, plan.Token)
	if err != nil {
		t.Fatal(err)
	}
	if result := waitJob(t, s, id, job.ID); result.State != "ready" {
		t.Fatal("lock survived cancellation", result)
	}
}
func TestArchiveValidationIntegrityAndCancelledExtraction(t *testing.T) {
	for _, name := range []string{"../escape", `C:\escape`, `dir\..\escape`, "/absolute", "file:stream", "NUL.txt", "COM1.dll", "folder/trailing.", "folder/trailing "} {
		if _, err := archiveName(name); err == nil {
			t.Fatal("accepted path", name)
		}
	}
	for _, raw := range []string{"http://buildbot.libretro.com/a", "https://buildbot.libretro.com.evil/a", "https://buildbot.libretro.com:8443/a", "https://user@buildbot.libretro.com/a", "https://raw.githubusercontent.com/other/repo/main/payload"} {
		if officialURL(raw) {
			t.Fatal("accepted URL", raw)
		}
	}
	dir := t.TempDir()
	data := zipBytes(t, map[string]string{"snes9x_libretro.dll": "core fixture"})
	path := filepath.Join(dir, "core.zip")
	writeFile(t, path, data)
	if err := extractCore(context.Background(), path, filepath.Join(dir, "out.dll"), "snes9x_libretro.dll"); err != nil {
		t.Fatal(err)
	}
	corrupt := append([]byte{}, data...)
	at := bytes.Index(corrupt, []byte("core fixture"))
	if at < 0 {
		t.Fatal("missing fixture bytes")
	}
	corrupt[at] ^= 0xff
	writeFile(t, path, corrupt)
	if err := extractCore(context.Background(), path, filepath.Join(dir, "bad.dll"), "snes9x_libretro.dll"); err == nil {
		t.Fatal("corrupt CRC accepted")
	}
	writeFile(t, path, zipBytes(t, map[string]string{"../escape.dll": "payload"}))
	if err := extractCore(context.Background(), path, filepath.Join(dir, "escape.dll"), "escape.dll"); err == nil {
		t.Fatal("zip traversal accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := extract7z(ctx, "testdata/sample.7z", filepath.Join(dir, "cancelled"), func(string) {}); err == nil {
		t.Fatal("ignored cancellation")
	}
	if err := extract7z(context.Background(), "testdata/sample.7z", filepath.Join(dir, "unpacked"), func(string) {}); err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "unpacked", "hello.txt"))
	if err != nil || !bytes.Equal(got, want) {
		t.Fatal(string(got), err)
	}
}

func TestOfficialEndpointsLive(t *testing.T) {
	if os.Getenv("GEMU_RETROARCH_LIVE_CHECK") != "1" {
		t.Skip("explicit opt-in: reads official release list and HEAD metadata only")
	}
	s := New(t.TempDir())
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	version, err := s.latest(ctx)
	if err != nil {
		t.Fatal(err)
	}
	urls := []string{buildbot + "/stable/" + version + "/windows/x86_64/RetroArch.7z"}
	seen := map[string]bool{}
	for _, sys := range Systems {
		if !seen[sys.Core] {
			seen[sys.Core] = true
			urls = append(urls, buildbot+"/nightly/windows/x86_64/latest/"+sys.Core+"_libretro.dll.zip", "https://raw.githubusercontent.com/libretro/libretro-core-info/master/"+sys.Core+"_libretro.info")
		}
	}
	for _, url := range urls {
		response, err := s.request(ctx, "HEAD", url)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
	}
	t.Log("official stable version:", version)
}

func TestOfficialPackageLive(t *testing.T) {
	if os.Getenv("GEMU_RETROARCH_PACKAGE_CHECK") != "1" {
		t.Skip("explicit opt-in: downloads and extracts official packages in a temporary directory; never executes RetroArch")
	}
	if !supported() {
		t.Skip("Windows x64 package")
	}
	s := New(t.TempDir())
	player, err := profiles.New(s.dir).Create("Package validation")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	plan, err := s.Plan(ctx, player.ID, InstallRequest{InstallApp: true, Systems: []string{"nes"}})
	if err != nil {
		t.Fatal(err)
	}
	job, err := s.Start(ctx, player.ID, plan.Token)
	if err != nil {
		t.Fatal(err)
	}
	for {
		current, err := s.Job(player.ID, job.ID)
		if err != nil {
			t.Fatal(err)
		}
		if terminal(current.State) {
			if current.State != "ready" {
				t.Fatal(current)
			}
			if err = s.validateExecutable(current.Executable); err != nil {
				t.Fatal(err)
			}
			if _, err = s.corePath(Config{Executable: current.Executable}, "fceumm"); err != nil {
				t.Fatal(err)
			}
			t.Log("validated official package:", plan.Version)
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(time.Second):
		}
	}
}

func TestRepairPreservesThePreviousManagedRelease(t *testing.T) {
	s, id, _, _ := fixture(t)
	s.client = &http.Client{Transport: packageTransport(t, false)}
	s.extract = fakeExtractor(t)
	originalCheck := s.checkPE
	s.checkPE = func(path string, dll bool) error {
		if err := originalCheck(path, dll); err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if string(data) == "damaged" {
			return errors.New("damaged executable")
		}
		return nil
	}
	release := filepath.Join(s.managedDir(), "releases", "1.22.2")
	writeFile(t, filepath.Join(release, "retroarch.exe"), []byte("damaged"))
	writeFile(t, filepath.Join(release, "gemu-install.json"), []byte(`{"version":"1.22.2"}`))
	plan, err := s.Plan(context.Background(), id, InstallRequest{InstallApp: true, Systems: []string{"snes"}})
	if err != nil {
		t.Fatal(err)
	}
	job, err := s.Start(context.Background(), id, plan.Token)
	if err != nil {
		t.Fatal(err)
	}
	if result := waitJob(t, s, id, job.ID); result.State != "ready" {
		t.Fatal(result)
	}
	backups, err := os.ReadDir(filepath.Join(s.managedDir(), "recovery"))
	if err != nil || len(backups) != 1 {
		t.Fatal(backups, err)
	}
	data, err := os.ReadFile(filepath.Join(s.managedDir(), "recovery", backups[0].Name(), "retroarch.exe"))
	if err != nil || string(data) != "damaged" {
		t.Fatal("damaged release was not retained", err)
	}
}
