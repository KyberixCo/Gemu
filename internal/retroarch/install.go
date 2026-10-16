package retroarch

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func New(dir string) *Service {
	if absolute, err := filepath.Abs(dir); err == nil {
		dir = absolute
	}
	return &Service{dir: dir, client: newHTTPClient(), plans: map[string]installPlan{}, jobs: map[string]*installJob{}, drafts: map[string]draft{}, extract: extract7z, checkPE: validatePE}
}
func (s *Service) Plan(ctx context.Context, id string, request InstallRequest) (Plan, error) {
	if err := safeProfileID(id); err != nil {
		return Plan{}, err
	}
	if !supported() {
		return Plan{}, fmt.Errorf("RetroArch setup is available on Windows x64")
	}
	if len(request.Systems) == 0 || len(request.Systems) > len(Systems) {
		return Plan{}, fmt.Errorf("select at least one RetroArch system")
	}
	if !request.InstallApp {
		if err := s.validateExecutable(request.Executable); err != nil {
			return Plan{}, err
		}
	}
	if request.CoreDirectory != "" {
		path, err := cleanDirectory(request.CoreDirectory)
		if err != nil {
			return Plan{}, err
		}
		request.CoreDirectory = path
	}
	plan := Plan{Token: rand.Text(), Destination: s.managedDir(), Downloads: []Download{}}
	if request.InstallApp {
		version, err := s.latest(ctx)
		if err != nil {
			return Plan{}, err
		}
		plan.Version = version
		existing := findExecutables(filepath.Join(s.managedDir(), "releases", version), 2)
		valid := false
		for _, path := range existing {
			if s.validateExecutable(path) == nil {
				valid = true
				break
			}
		}
		if !valid {
			plan.Downloads = append(plan.Downloads, Download{Name: "RetroArch.7z", URL: buildbot + "/stable/" + version + "/windows/x86_64/RetroArch.7z"})
		}
	}
	seen := map[string]bool{}
	for _, id := range request.Systems {
		sys, ok := system(id)
		if !ok {
			return Plan{}, fmt.Errorf("unknown RetroArch system")
		}
		if seen[sys.Core] {
			continue
		}
		seen[sys.Core] = true
		if _, err := s.corePath(Config{Executable: request.Executable, CoreDirectory: request.CoreDirectory}, sys.Core); err == nil {
			continue
		}
		name := sys.Core + "_libretro.dll.zip"
		plan.Downloads = append(plan.Downloads, Download{Name: name, URL: buildbot + "/nightly/windows/x86_64/latest/" + name}, Download{Name: sys.Core + "_libretro.info", URL: "https://raw.githubusercontent.com/libretro/libretro-core-info/master/" + sys.Core + "_libretro.info"})
	}
	for i, asset := range plan.Downloads {
		headCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		response, err := s.request(headCtx, "HEAD", asset.URL)
		cancel()
		if err != nil {
			return Plan{}, err
		}
		plan.Downloads[i].Size = response.ContentLength
		response.Body.Close()
	}
	plan.RecommendedBytes = 64 * 1024 * 1024
	for _, asset := range plan.Downloads {
		size := asset.Size
		if size < 0 {
			size = 256 * 1024 * 1024
		}
		plan.RecommendedBytes += size
		if asset.Name == "RetroArch.7z" {
			plan.RecommendedBytes += 1024 * 1024 * 1024
		}
		if strings.HasSuffix(asset.Name, ".dll.zip") {
			plan.RecommendedBytes += 256 * 1024 * 1024
		}
	}
	plan.FreeBytes = -1
	if available, e := freeSpace(plan.Destination); e == nil {
		plan.FreeBytes = available
	}
	if plan.FreeBytes >= 0 && plan.FreeBytes < plan.RecommendedBytes {
		return Plan{}, fmt.Errorf("insufficient free space for RetroArch installation")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for token, p := range s.plans {
		if time.Now().After(p.expires) {
			delete(s.plans, token)
		}
	}
	if len(s.plans) >= 8 {
		return Plan{}, fmt.Errorf("too many pending RetroArch installation plans")
	}
	s.plans[plan.Token] = installPlan{plan: plan, owner: id, request: request, expires: time.Now().Add(15 * time.Minute)}
	return plan, nil
}
func terminal(state string) bool { return state == "ready" || state == "error" || state == "cancelled" }
func (s *Service) Start(parent context.Context, id, token string) (Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	plan, ok := s.plans[token]
	if !ok || plan.owner != id || time.Now().After(plan.expires) {
		return Job{}, fmt.Errorf("RetroArch installation plan expired; review it again")
	}
	if s.busy {
		return Job{}, fmt.Errorf("another RetroArch installation is already running")
	}
	if len(s.jobs) >= 8 {
		for key, job := range s.jobs {
			if terminal(job.view.State) {
				delete(s.jobs, key)
				break
			}
		}
	}
	if len(s.jobs) >= 8 {
		return Job{}, fmt.Errorf("too many RetroArch installation jobs")
	}
	ctx, cancel := context.WithTimeout(parent, 30*time.Minute)
	view := Job{ID: rand.Text(), State: "starting", Count: len(plan.plan.Downloads)}
	job := &installJob{view: view, owner: id, cancel: cancel}
	s.jobs[view.ID] = job
	s.busy = true
	delete(s.plans, token)
	go func() {
		defer cancel()
		exe, err := s.install(ctx, plan, func(update func(*Job)) { s.mu.Lock(); defer s.mu.Unlock(); update(&job.view) })
		s.mu.Lock()
		defer s.mu.Unlock()
		s.busy = false
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				job.view.State = "cancelled"
			} else {
				job.view.State = "error"
				job.view.Error = err.Error()
			}
		} else {
			job.view.State = "ready"
			job.view.Executable = exe
		}
	}()
	return view, nil
}
func (s *Service) Job(id, key string) (Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job := s.jobs[key]
	if job == nil || job.owner != id {
		return Job{}, fmt.Errorf("RetroArch installation job is not available")
	}
	return job.view, nil
}
func (s *Service) Cancel(id, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	job := s.jobs[key]
	if job == nil || job.owner != id {
		return fmt.Errorf("RetroArch installation job is not available")
	}
	if job.view.State == "publishing" {
		return fmt.Errorf("RetroArch installation is finishing")
	}
	if !terminal(job.view.State) {
		job.cancel()
	}
	return nil
}

func (s *Service) DiscardPlan(id, token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p, ok := s.plans[token]; ok && p.owner == id {
		delete(s.plans, token)
	}
}

func (s *Service) install(ctx context.Context, plan installPlan, update func(func(*Job))) (string, error) {
	root := s.managedDir()
	if err := os.MkdirAll(root, 0700); err != nil {
		return "", err
	}
	release, err := installLock(root)
	if err != nil {
		return "", err
	}
	defer release()
	// Only our private staging names inside this root are cleaned, and only
	// while holding the cross-process installation lock. Live releases survive.
	r, err := os.OpenRoot(root)
	if err != nil {
		return "", err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		r.Close()
		return "", err
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".stage-") {
			if err = r.RemoveAll(entry.Name()); err != nil {
				r.Close()
				return "", err
			}
		}
	}
	r.Close()
	stage, err := os.MkdirTemp(root, ".stage-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(stage)
	for _, sub := range []string{"downloads", "app", "cores", "info"} {
		if err = os.MkdirAll(filepath.Join(stage, sub), 0700); err != nil {
			return "", err
		}
	}
	digests := map[string]string{}
	var unpackedExe string
	for i, asset := range plan.plan.Downloads {
		update(func(j *Job) {
			j.State = "downloading"
			j.Current = asset.Name
			j.Bytes = 0
			j.Total = asset.Size
			j.Completed = i
		})
		path := filepath.Join(stage, "downloads", asset.Name)
		digest, e := s.download(ctx, asset, path, func(read, total int64) { update(func(j *Job) { j.Bytes = read; j.Total = total }) })
		if e != nil {
			return "", e
		}
		digests[asset.Name] = digest
		update(func(j *Job) { j.State = "extracting"; j.Current = asset.Name })
		switch {
		case asset.Name == "RetroArch.7z":
			if err = s.extract(ctx, path, filepath.Join(stage, "app"), func(name string) { update(func(j *Job) { j.Current = name }) }); err != nil {
				return "", err
			}
			found := findExecutables(filepath.Join(stage, "app"), 2)
			if len(found) != 1 {
				return "", fmt.Errorf("official RetroArch package does not contain a unique executable")
			}
			unpackedExe = found[0]
			if err = s.validateExecutable(unpackedExe); err != nil {
				return "", err
			}
		case strings.HasSuffix(asset.Name, ".dll.zip"):
			name := strings.TrimSuffix(asset.Name, ".zip")
			out := filepath.Join(stage, "cores", name)
			if err = extractCore(ctx, path, out, name); err != nil {
				return "", err
			}
			if err = s.checkPE(out, true); err != nil {
				return "", err
			}
		case strings.HasSuffix(asset.Name, ".info"):
			data, e := os.ReadFile(path)
			if e != nil {
				return "", e
			}
			if len(data) > 1024*1024 || !strings.Contains(string(data), "display_name") {
				return "", fmt.Errorf("invalid RetroArch core information")
			}
			if err = os.Rename(path, filepath.Join(stage, "info", asset.Name)); err != nil {
				return "", err
			}
		default:
			return "", fmt.Errorf("unknown RetroArch installation asset")
		}
	}
	if err = ctx.Err(); err != nil {
		return "", err
	}
	update(func(j *Job) { j.State = "publishing"; j.Current = "" })
	executable := plan.request.Executable
	if plan.request.InstallApp {
		destination := filepath.Join(root, "releases", plan.plan.Version)
		if unpackedExe != "" {
			rel, _ := filepath.Rel(filepath.Join(stage, "app"), unpackedExe)
			receipt, _ := json.MarshalIndent(map[string]any{"version": plan.plan.Version, "downloads": plan.plan.Downloads, "sha256": digests}, "", "  ")
			if err = os.WriteFile(filepath.Join(stage, "app", "gemu-install.json"), receipt, 0600); err != nil {
				return "", err
			}
			if err = os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
				return "", err
			}
			reuse := ""
			backup := ""
			if _, e := os.Stat(destination); e == nil {
				for _, path := range findExecutables(destination, 2) {
					if s.validateExecutable(path) == nil {
						reuse = path
						break
					}
				}
				if reuse == "" {
					// Repair only a release with our own receipt, preserving the
					// damaged copy. Unrecognized directories are never replaced.
					marker, e := os.ReadFile(filepath.Join(destination, "gemu-install.json"))
					var previous struct {
						Version string `json:"version"`
					}
					if e != nil || len(marker) > 1024*1024 || json.Unmarshal(marker, &previous) != nil || previous.Version != plan.plan.Version {
						return "", fmt.Errorf("RetroArch release directory already exists; select it manually or repair its files")
					}
					backup = filepath.Join(root, "recovery", plan.plan.Version+"-"+rand.Text())
					if err = os.MkdirAll(filepath.Dir(backup), 0700); err != nil {
						return "", err
					}
					if err = os.Rename(destination, backup); err != nil {
						return "", err
					}
				}
			}
			if reuse != "" {
				executable = reuse
			} else {
				if err = os.Rename(filepath.Join(stage, "app"), destination); err != nil {
					if backup != "" {
						_ = os.Rename(backup, destination)
					}
					return "", err
				}
				executable = filepath.Join(destination, rel)
			}
		} else {
			for _, path := range findExecutables(destination, 2) {
				if s.validateExecutable(path) == nil {
					executable = path
					break
				}
			}
		}
	}
	for _, sub := range []string{"cores", "info"} {
		if err = os.MkdirAll(filepath.Join(root, sub), 0700); err != nil {
			return "", err
		}
		files, e := os.ReadDir(filepath.Join(stage, sub))
		if e != nil {
			return "", e
		}
		for _, file := range files {
			if err = os.Rename(filepath.Join(stage, sub, file.Name()), filepath.Join(root, sub, file.Name())); err != nil {
				return "", err
			}
		}
	}
	if err = s.validateExecutable(executable); err != nil {
		return "", err
	}
	for _, id := range plan.request.Systems {
		sys, _ := system(id)
		if _, err = s.corePath(Config{Executable: executable, CoreDirectory: plan.request.CoreDirectory}, sys.Core); err != nil {
			return "", err
		}
	}
	update(func(j *Job) { j.Completed = j.Count })
	return executable, nil
}

// Read a small metadata response without trusting its declared content length.
func readSmall(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, 1024*1024+1))
	if len(data) > 1024*1024 {
		return nil, fmt.Errorf("RetroArch metadata is too large")
	}
	return data, err
}
