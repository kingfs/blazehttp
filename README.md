# BlazeHTTP

一个非标准http协议解析库, 弥补标准库对非标准请求的解析，常用于安全测试等目的，并且非常快!

## 帮助

```bash
go get github.com/kingfs/blazehttp
```

## 测试

```bash
# 构建测试工具
go build ./cmd/blazehttp

# 测试请求
./blazehttp -t http://192.168.0.1:8080 -g './testcases/*.http'
```
