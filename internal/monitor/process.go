package monitor

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"carbonqt/internal/models"

	"github.com/shirou/gopsutil/v3/process"
)

func ListProcesses(repoRoot string) ([]models.ProcessMetrics, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	results := make([]models.ProcessMetrics, 0, len(procs))
	for _, proc := range procs {
		metrics, ok := collectProcess(proc, repoRoot)
		if !ok {
			continue
		}
		results = append(results, metrics)
	}

	return results, nil
}

func KillProcess(pid int32) error {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}

func collectProcess(proc *process.Process, repoRoot string) (models.ProcessMetrics, bool) {
	name := resolveProcessName(proc)
	if name == "" {
		return models.ProcessMetrics{}, false
	}

	cpuPercent, err := proc.CPUPercent()
	if err != nil {
		cpuPercent = 0
	}

	memPercent, err := proc.MemoryPercent()
	if err != nil {
		memPercent = 0
	}

	var ctxSwitches *uint64
	if ctx, err := proc.NumCtxSwitches(); err == nil && ctx != nil {
		value := uint64(ctx.Voluntary + ctx.Involuntary)
		ctxSwitches = &value
	}

	cwd := ""
	if wd, err := proc.Cwd(); err == nil {
		cwd = wd
	}

	if repoRoot != "" {
		if cwd == "" || !isInRepo(cwd, repoRoot) {
			return models.ProcessMetrics{}, false
		}
	}

	exePath := ""
	if exe, err := proc.Exe(); err == nil {
		exePath = strings.TrimSpace(exe)
	}

	startTime := time.Time{}
	if created, err := proc.CreateTime(); err == nil && created > 0 {
		startTime = time.UnixMilli(created)
	}

	return models.ProcessMetrics{
		PID:             proc.Pid,
		Name:            name,
		CPUPercent:      cpuPercent,
		MemoryPercent:   float64(memPercent),
		ContextSwitches: ctxSwitches,
		Cwd:             cwd,
		ExePath:         exePath,
		StartTime:       startTime,
	}, true
}

func resolveProcessName(proc *process.Process) string {
	if cmdline, err := proc.Cmdline(); err == nil {
		cmdline = strings.TrimSpace(cmdline)
		if cmdline != "" {
			return parseExecutable(cmdline)
		}
	}

	if exe, err := proc.Exe(); err == nil {
		exe = strings.TrimSpace(exe)
		if exe != "" {
			return filepath.Base(exe)
		}
	}

	if name, err := proc.Name(); err == nil {
		return strings.TrimSpace(name)
	}

	return ""
}

func parseExecutable(cmdline string) string {
	if cmdline == "" {
		return ""
	}

	if cmdline[0] == '"' {
		end := strings.Index(cmdline[1:], "\"")
		if end >= 0 {
			return filepath.Base(cmdline[1 : end+1])
		}
	}

	fields := strings.Fields(cmdline)
	if len(fields) == 0 {
		return ""
	}

	return filepath.Base(fields[0])
}

func isInRepo(cwd, repoRoot string) bool {
	cleanCwd := filepath.Clean(cwd)
	cleanRoot := filepath.Clean(repoRoot)

	if runtime.GOOS == "windows" {
		cleanCwd = strings.ToLower(cleanCwd)
		cleanRoot = strings.ToLower(cleanRoot)
	}

	if cleanCwd == cleanRoot {
		return true
	}

	separator := string(os.PathSeparator)
	return strings.HasPrefix(cleanCwd, cleanRoot+separator)
}
