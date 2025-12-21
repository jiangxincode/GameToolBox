---
title: 从源码构建
---

# 从源码构建

## Windows

```powershell
go mod download
go build -o game_tool_box.exe .\cmd\game_tool_box
```

## Linux / macOS

Fyne 使用 GLFW，通常需要系统图形依赖与 CGO（CI 里也会安装相关依赖）。

```bash
go mod download
go build -o game_tool_box ./cmd/game_tool_box
```

