# xfcap

手机 Root 上按 UID 做 iptables `REDIRECT`，本机小转发器转到电脑 mitmproxy / Reqable。不建 tun。

```bash
make android
xfcap start -p com.zhihu.android -h 172.18.22.226:9000
xfcap start --pid 4321 -h 172.18.22.226:9000
xfcap status
```

`-p` 会把这个包的主 UID、`:push`、isolatedProcess 一并写进 iptables。`--pid` 只解析成 UID，规则仍按 UID。
