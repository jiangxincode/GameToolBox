## How To Contribute

* Report a bug or request a new feature by [Open an issue](https://github.com/jiangxincode/GameToolBox/issues/new)

* Submit Pull Request if you are interested in development with us:
    * Fix some bugs
    * Add some new features
    * Improve the documents
    * Refactor the code
    * Fix some bad code smell reported by various CI tools
    * Optimize the performance
    * Add some tests
    * Add some translations

* Help us to broadcast this project and let more people know it

Feel free to dive in! You can do everything to help this project. All you need to do is to follow the [Contributor Covenant](http://contributor-covenant.org/version/1/3/0/).

### How to build the project

1. 安装Go(1.20+)，并配置好环境变量 `GOPATH` 和 `GOROOT`。
2. 下载依赖：`go mod download`
3. 构建：`go build -o game_tool_box.exe .\cmd\game_tool_box`
4. 运行：`./game_tool_box.exe`

## 项目结构

- `cmd/game_tool_box/`：正式应用入口（Fyne GUI）。
- `internal/resources/`：内嵌资源（例如窗口图标）。
- `examples/`：历史/练习用的独立示例程序（可忽略）。

### Contributors

* Jiangxin <jiangxinnju@gmail.com>
