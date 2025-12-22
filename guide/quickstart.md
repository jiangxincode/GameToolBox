## 快速开始
下载即编译：
```bash
go install github.com/yourname/your-go-project/cmd/cli@latest
```
一行代码调用：
```go
package main

import "github.com/yourname/your-go-project"

func main() {
    client := myapp.New("your-api-key")
    client.Run()
}
```