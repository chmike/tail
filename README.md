# Tail

[![GitHub Release](https://img.shields.io/github/v/release/chmike/tail)](https://github.com/chmike/tail/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/chmike/tail.svg)](https://pkg.go.dev/github.com/chmike/tail)
[![go.mod](https://img.shields.io/github/go-mod/go-version/chmike/tail)](go.mod)
[![Build Status](https://img.shields.io/github/workflow/status/chmike/tail/build)](https://github.com/chmike/tail/actions?query=workflow%3Abuild+branch%3Amaster)
[![Go Report Card](https://goreportcard.com/badge/github.com/chmike/tail)](https://goreportcard.com/report/github.com/chmike/tail)
[![Codecov](https://codecov.io/gh/chmike/tail/branch/main/graph/badge.svg)](https://codecov.io/gh/chmike/tail)

A go module/package to read all lines appended to a file at runtime. It support file rotation which is a common practice for log files. It currently works on Linux and MacOS, not on Windows. 

```
import "github.com/chmike/tail"

func main() {
  tail := NewTail("/var/log/auth.log")
  for {
    select {
      case line := <-tail.Line:
        println(line)
      case err := <-tail.Error:
        print("error:")
        println(err.Error())
        break
    }
  }
}
```
