## 通过何种方式参与项目

欢迎加入！任何形式的帮助都非常感谢：

- 通过 [提交 Issue](https://github.com/jiangxincode/GameToolBox/issues/new) 来报告 Bug 或提出新功能需求。
- 如果你愿意参与开发，欢迎提交 Pull Request（PR），你可以：
  - 修复一些 Bug
  - 添加新功能
  - 改进文档
  - 重构代码
  - 修复各类 CI 工具报告的代码异味
  - 优化性能
  - 补充测试
  - 添加翻译
- 帮助宣传这个项目，让更多人知道它。

参与贡献的基本行为规范请参考 [Contributor Covenant](http://contributor-covenant.org/version/1/3/0/)。

## 如何构建项目

1. 安装 Go（1.20+），并配置好环境变量 `GOPATH` 和 `GOROOT`。
2. 下载依赖：`go mod download`
3. 构建：
    1. Windows: `go build -o game_tool_box.exe .\cmd\game_tool_box`
    2. macOS: `go build -o game_tool_box ./cmd/game_tool_box`
    3. Linux: `sudo apt-get install -y libgl1-mesa-dev xorg-dev; go build -o game_tool_box ./cmd/game_tool_box`
4. 直接运行生成的`game_tool_box`程序即可。

## 项目结构

- `cmd/game_tool_box/`：应用入口（Fyne GUI）。
- `internal/ui/`：UI 相关代码。
- `internal/config/`：配置相关。
- `internal/log/`：日志相关。
- `internal/i18n/`：国际化相关。
- `internal/resources/`：内嵌资源。
- `docs/`：项目文档。

## 贡献者

- Jiangxin <jiangxinnju@gmail.com>
