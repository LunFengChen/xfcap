package main

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/LunFengChen/xfcap/internal/app"
	"github.com/LunFengChen/xfcap/internal/capture"
	"github.com/LunFengChen/xfcap/internal/fw"
	"github.com/LunFengChen/xfcap/internal/redir"
)

func start(a *args) {
	if os.Getenv(envDaemon) == "1" {
		runDaemon(a)
		return
	}
	if err := a.requireCapture(); err != nil {
		fail(err.Error())
	}
	add, err := itemsFromArgs(a)
	if err != nil {
		fail(err.Error())
	}
	if err := os.MkdirAll(runDir(a.home), 0755); err != nil {
		fail(err.Error())
	}
	// 再 start 一次是追加 App，同包名则覆盖转到哪。
	cs := capture.Add(capture.Load(a.home), add)
	if err := capture.Save(a.home, cs); err != nil {
		fail(err.Error())
	}
	gen, err := bumpGen(a.home)
	if err != nil {
		fail(err.Error())
	}
	if pid := alivePID(a.home); pid > 0 {
		// 守护进程已在听 17892，SIGHUP 只重装 iptables / 路由，不断连接。
		if err := syscall.Kill(pid, syscall.SIGHUP); err != nil {
			fail(err.Error())
		}
		if !waitApplied(a.home, gen, 8*time.Second) {
			fail("reload timeout")
		}
		os.Stdout.WriteString(capture.Table(cs))
		return
	}
	if err := spawnDaemon(a.home); err != nil {
		fail(err.Error())
	}
	if !waitApplied(a.home, gen, 8*time.Second) {
		fail("daemon timeout")
	}
}

func runDaemon(a *args) {
	if !fw.HasOwnerMatch() {
		fail("iptables owner match missing")
	}
	srv := &redir.Server{}
	reload := func() error {
		cs := capture.Load(a.home)
		if len(cs) == 0 {
			return fmt.Errorf("no apps")
		}
		cs = refreshUIDs(cs)
		if err := capture.Save(a.home, cs); err != nil {
			return err
		}
		byUID := map[int]redir.Upstream{}
		var all *redir.Upstream
		for _, c := range cs {
			u, err := redir.Parse(c.Host)
			if err != nil {
				return err
			}
			if c.All {
				cp := u
				all = &cp
				continue
			}
			for _, id := range c.AllUIDs() {
				byUID[id] = u
			}
		}
		srv.SetRoutes(byUID, all)
		if err := srv.Listen(a.redirPort); err != nil {
			return err
		}
		if err := fw.Install(specFrom(cs, a.redirPort)); err != nil {
			return err
		}
		markApplied(a.home)
		return nil
	}
	ch := make(chan os.Signal, 8)
	waitSignals(ch)
	if err := reload(); err != nil {
		fail(err.Error())
	}
	_ = writePID(a.home)
	_ = writeState(a.home, capture.Table(capture.Load(a.home)))
	tick := time.NewTicker(15 * time.Second)
	defer tick.Stop()
	for {
		select {
		case sig := <-ch:
			if sig == syscall.SIGHUP {
				if err := reload(); err != nil {
					fmt.Fprintln(os.Stderr, "reload:", err)
					continue
				}
				_ = writeState(a.home, capture.Table(capture.Load(a.home)))
				continue
			}
			fw.Cleanup()
			srv.Close()
			return
		case <-tick.C:
			// isolated / :push 后起，重扫 UID 再装规则。
			if err := reload(); err != nil {
				fmt.Fprintln(os.Stderr, "reload:", err)
				continue
			}
			_ = writeState(a.home, capture.Table(capture.Load(a.home)))
		}
	}
}

func itemsFromArgs(a *args) ([]capture.Item, error) {
	var out []capture.Item
	if a.all {
		out = append(out, capture.Item{All: true, Host: a.upstream})
	}
	for _, p := range a.packages {
		us, err := app.UIDs(p)
		if err != nil {
			return nil, err
		}
		out = append(out, capture.Item{Pkg: p, UID: us[0], UIDs: us, Host: a.upstream})
	}
	for _, u := range a.uids {
		if u > 0 {
			out = append(out, capture.Item{UID: u, UIDs: []int{u}, Host: a.upstream})
		}
	}
	for _, pid := range a.pids {
		u, err := app.FromPID(pid)
		if err != nil {
			return nil, err
		}
		out = append(out, capture.Item{UID: u, UIDs: []int{u}, Host: a.upstream})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("need -p PKG")
	}
	return out, nil
}

func specFrom(cs []capture.Item, redirPort int) fw.Spec {
	s := fw.Spec{Redir: redirPort}
	seen := map[int]struct{}{}
	for _, c := range cs {
		if c.All {
			s.All = true
		}
		for _, id := range c.AllUIDs() {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			s.UIDs = append(s.UIDs, id)
		}
	}
	return s
}

func refreshUIDs(cs []capture.Item) []capture.Item {
	out := make([]capture.Item, len(cs))
	copy(out, cs)
	for i, c := range out {
		if c.All || c.Pkg == "" {
			continue
		}
		us, err := app.UIDs(c.Pkg)
		if err != nil || len(us) == 0 {
			continue
		}
		us = app.Merge(c.AllUIDs(), us)
		out[i].UIDs = us
		out[i].UID = us[0]
	}
	return out
}
