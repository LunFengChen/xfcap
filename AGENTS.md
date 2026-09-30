# xfcap

## 为什么用

正常抓包走系统 VPN，很多 App 会检测。xfcap 不建 tun：在已经是 root 的手机上，按 UID 把 App 的 TCP `REDIRECT` 到本机小转发器，读原来的目的地址，再 HTTP CONNECT / SOCKS5 到电脑 mitm（mitmproxy、Reqable）。

仓库：`/home/xiaofeng/Desktop/projects/xfcap`。手机上跑 `dist/xfcap`。

## 怎么用

```bash
xfcap start -p 包名 -h 主机:端口
```

再抓一个 App 再 `start` 一次，会加进去，不会把前面的顶掉。`-p` 会带上这个包的 `:push` / isolated UID。`--pid` 先查进程 UID，规则仍按 UID。

```bash
xfcap start -p com.zhihu.android -h 172.18.22.226:9000
xfcap start -p com.android.chrome -h 172.18.22.226:9000
xfcap status
```

`status` 是一张表：包名、UID、转到哪。

## 参数

| 命令 | 作用 |
|---|---|
| `start -p PKG -h HOST:PORT` | 加上这个 App（含 :push / isolated 的 UID） |
| `start --pid PID -h HOST:PORT` | 按进程查 UID 再加 |
| `stop` | 全部停 |
| `status` | 正在抓哪些 |
| `doctor` | iptables / root / 守护进程 |
| `uid PKG` | 查这个包全部 UID |
| `uid --pid PID` | 进程 → UID |

| 参数 | 作用 |
|---|---|
| `-p` / `--package` | 包名，可多次 |
| `-h` / `--host` | 电脑 mitm，`IP:PORT` 或 `socks5://IP:PORT` |
| `--uid N` | 直接指定 UID |
| `--pid N` | 进程 PID，内部转成 UID（iptables 仍按 UID） |
| `--all` | 所有 App |
| `--help` | 帮助 |

`status` 的 UID 列可能是 `10175,99001`：同一包多个 UID。守护进程每 15 秒重扫一次，后起的 isolated 进程会补进规则。
