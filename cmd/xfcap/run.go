package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// 前台 start 只改列表、拉起或通知守护进程；真正听端口的是 XFCAP_DAEMON=1 那份。
const envDaemon = "XFCAP_DAEMON"

func runDir(home string) string    { return filepath.Join(home, "run") }
func pidPath(home string) string      { return filepath.Join(runDir(home), "xfcap.pid") }
func statePath(home string) string    { return filepath.Join(runDir(home), "state.txt") }
func logPath(home string) string      { return filepath.Join(runDir(home), "xfcap.log") }
func genPath(home string) string      { return filepath.Join(runDir(home), "gen.txt") }
func appliedPath(home string) string  { return filepath.Join(runDir(home), "applied.txt") }

func writePID(home string) error {
	return os.WriteFile(pidPath(home), []byte(strconv.Itoa(os.Getpid())), 0644)
}

func alivePID(home string) int {
	b, err := os.ReadFile(pidPath(home))
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || pid <= 0 {
		return 0
	}
	if err := syscall.Kill(pid, 0); err != nil {
		return 0
	}
	return pid
}

func killDaemon(home string) {
	b, _ := os.ReadFile(pidPath(home))
	pid, _ := strconv.Atoi(strings.TrimSpace(string(b)))
	if pid > 0 && pid != os.Getpid() {
		_ = syscall.Kill(pid, syscall.SIGTERM)
		time.Sleep(300 * time.Millisecond)
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
	_ = os.Remove(statePath(home))
	_ = os.Remove(pidPath(home))
	_ = os.Remove(appliedPath(home))
}

func writeState(home, text string) error {
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return os.WriteFile(statePath(home), []byte(text), 0644)
}

func spawnDaemon(home string) error {
	if err := os.MkdirAll(runDir(home), 0755); err != nil {
		return err
	}
	logf, err := os.Create(logPath(home))
	if err != nil {
		return err
	}
	cmd := exec.Command("/proc/self/exe", os.Args[1:]...)
	cmd.Env = append(os.Environ(), envDaemon+"=1")
	cmd.Stdout = logf
	cmd.Stderr = logf
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true} // 脱离 adb shell，关掉终端也不死
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("daemon: %w", err)
	}
	go cmd.Wait()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if b, err := os.ReadFile(statePath(home)); err == nil && len(b) > 0 {
			os.Stdout.Write(b)
			if b[len(b)-1] != '\n' {
				os.Stdout.Write([]byte("\n"))
			}
			return nil
		}
		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			tail, _ := os.ReadFile(logPath(home))
			return fmt.Errorf("daemon exited: %s", firstLine(string(tail)))
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("daemon timeout, see %s", logPath(home))
}

func waitSignals(ch chan os.Signal) {
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
}

func bumpGen(home string) (int, error) {
	n := 1
	if b, err := os.ReadFile(genPath(home)); err == nil {
		if v, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil && v > 0 {
			n = v + 1
		}
	}
	if err := os.WriteFile(genPath(home), []byte(strconv.Itoa(n)+"\n"), 0644); err != nil {
		return 0, err
	}
	return n, nil
}

func markApplied(home string) {
	b, err := os.ReadFile(genPath(home))
	if err != nil {
		return
	}
	_ = os.WriteFile(appliedPath(home), b, 0644)
}

func waitApplied(home string, gen int, d time.Duration) bool {
	if gen <= 0 {
		return true
	}
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		b, err := os.ReadFile(appliedPath(home))
		if err == nil {
			v, _ := strconv.Atoi(strings.TrimSpace(string(b)))
			if v >= gen {
				return true
			}
		}
		if alivePID(home) == 0 {
			return false
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
