# GameToolBox

[![Apache License V2.0](https://img.shields.io/badge/license-Apache%202-green)](http://www.apache.org/licenses/LICENSE-2.0)
[![GitHub all releases](https://img.shields.io/github/downloads/jiangxincode/GameToolBox/total)](https://github.com/jiangxincode/GameToolBox/releases)
[![Build and Rel
ease](https://github.com/jiangxincode/GameToolBox/actions/workflows/BuildAndRelease.yml/badge.svg)](https://github.com/jiangxincode/GameToolBox/actions/workflows/BuildAndRelease.yml)

[Deutsch](https://www.readme-i18n.com/jiangxincode/GameToolBox?lang=de) |
[Español](https://www.readme-i18n.com/jiangxincode/GameToolBox?lang=es) |
[français](https://www.readme-i18n.com/jiangxincode/GameToolBox?lang=fr) |
[日本語](https://www.readme-i18n.com/jiangxincode/GameToolBox?lang=ja) |
[한국어](https://www.readme-i18n.com/jiangxincode/GameToolBox?lang=ko) |
[Português](https://www.readme-i18n.com/jiangxincode/GameToolBox?lang=pt) |
[Русский](https://www.readme-i18n.com/jiangxincode/GameToolBox?lang=ru) |
[中文](https://www.readme-i18n.com/jiangxincode/GameToolBox?lang=zh)


一个使用 Go 语言编写的游戏整理工具箱，使用 Fyne 作为 GUI。

## 功能（规划/进行中）

- 解析 天马G/PEGASUS 的 `metadata.pegasus.txt`，并自动生成空的游戏 ROM 文件以便进行测试。
- 解析 天马G/PEGASUS 的 `metadata.pegasus.txt`，并自动生成 `media/` 文件夹下的所有子目录。
- 将 天马G/PEGASUS 的游戏列表转换为 Batocera 的 `gamelist.xml` 格式。

## 构建与运行

1. 安装Go(1.20+)，并配置好环境变量 `GOPATH` 和 `GOROOT`。
2. 下载依赖：`go mod download`
3. 构建：`go build -o game_tool_box.exe .\cmd\game_tool_box`
4. 运行：`./game_tool_box.exe`

## 项目结构

- `cmd/game_tool_box/`：正式应用入口（Fyne GUI）。
- `internal/resources/`：内嵌资源（例如窗口图标）。
- `examples/`：历史/练习用的独立示例程序（可忽略）。



