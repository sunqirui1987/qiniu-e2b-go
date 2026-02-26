# Qiniu E2B Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/qiniu/e2b-go.svg)](https://pkg.go.dev/github.com/qiniu/e2b-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/qiniu/e2b-go)](https://goreportcard.com/report/github.com/qiniu/e2b-go)

Go SDK for Qiniu Cloud E2B Sandbox - 在云端安全沙箱中执行代码和管理文件系统。

## 特性

- 代码执行支持多种语言（Python, JavaScript, TypeScript, Bash, Go, Rust, Java）
- 文件系统操作（读取、写入、列表、删除、创建目录）
- 流式输出处理（NDJSON 格式）
- Context 上下文管理
- 图表数据支持（Line, Scatter, Bar, Pie, BoxAndWhisker, SuperChart）

## 安装

```bash
go get github.com/qiniu/e2b-go
```

## 快速开始

### 基础用法

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    e2b "github.com/qiniu/e2b-go"
)

func main() {
    // 设置环境变量
    os.Setenv("E2B_API_KEY", "your-api-key")
    os.Setenv("E2B_API_URL", "https://cn-yangzhou-1-sandbox.qiniuapi.com")

    ctx := context.Background()

    // 创建沙箱
    sbx, err := e2b.Create(ctx, &e2b.SandboxOpts{
        Template:  "code-interpreter-v1",
        TimeoutMs: 300000, // 5 分钟
    })
    if err != nil {
        log.Fatalf("创建沙箱失败: %v", err)
    }
    defer sbx.Kill() // 确保沙箱被清理

    fmt.Printf("沙箱已创建: %s\n", sbx.SandboxID())

    // 执行 Python 代码
    execution, err := sbx.RunCode(`print("Hello, World!")`, &e2b.RunCodeOpts{
        Language: e2b.Python,
    })
    if err != nil {
        log.Fatalf("执行代码失败: %v", err)
    }

    // 打印输出
    for _, log := range execution.Logs {
        fmt.Println(log.Line)
    }
}
```

### 执行 JavaScript 代码

```go
// 执行 JavaScript 代码
jsExecution, err := sbx.RunCode(`
const sum = (a, b) => a + b;
console.log("1 + 2 =", sum(1, 2));
for (let i = 0; i < 3; i++) {
    console.log("iteration", i);
}
`, &e2b.RunCodeOpts{
    Language: e2b.JavaScript,
})
if err != nil {
    log.Fatalf("执行失败: %v", err)
}

for _, log := range jsExecution.Logs {
    fmt.Print(log.Line)
}
```

### 文件系统操作

```go
// 写入文件
err := sbx.Files.Write("/home/user/test.txt", []byte("Hello, File!"))
if err != nil {
    log.Fatalf("写入文件失败: %v", err)
}

// 读取文件
content, err := sbx.Files.Read("/home/user/test.txt")
if err != nil {
    log.Fatalf("读取文件失败: %v", err)
}
fmt.Println("文件内容:", content)

// 列出目录
entries, err := sbx.Files.List("/home/user")
if err != nil {
    log.Fatalf("列出目录失败: %v", err)
}
for _, entry := range entries {
    fmt.Printf("%s (%s) - %d bytes\n", entry.Name, entry.Type, entry.Size)
}

// 创建目录
err = sbx.Files.MakeDir("/home/user/mydir")

// 检查文件是否存在
exists, err := sbx.Files.Exists("/home/user/test.txt")

// 删除文件
err = sbx.Files.Remove("/home/user/test.txt")
```

## API 参考

### 创建沙箱

```go
sbx, err := e2b.Create(ctx, &e2b.SandboxOpts{
    Template:  "code-interpreter-v1", // 模板 ID
    TimeoutMs: 300000,                // 超时时间（毫秒）
    EnvVars: map[string]string{       // 环境变量
        "FOO": "bar",
    },
    Metadata: map[string]string{},    // 元数据
})
```

### 执行代码

```go
execution, err := sbx.RunCode(code, &e2b.RunCodeOpts{
    Language:  e2b.Python,            // 语言
    TimeoutMs: 60000,                 // 超时时间
    ContextID: "",                    // 上下文 ID（可选）
    EnvVars:   map[string]string{},   // 环境变量（可选）
    OnStdout:  func(msg *OutputMessage) { // stdout 回调
        fmt.Println(msg.Line)
    },
    OnStderr: func(msg *OutputMessage) { // stderr 回调
        fmt.Fprintln(os.Stderr, msg.Line)
    },
    OnResult: func(result *Result) {  // 结果回调
        fmt.Println(result.Text)
    },
    OnError: func(err *ExecutionError) { // 错误回调
        fmt.Printf("Error: %s - %s\n", err.Name, err.Value)
    },
})
```

### 支持的语言

| 语言 | 常量 | 运行时 |
|------|------|--------|
| Python | `e2b.Python` | python3 |
| JavaScript | `e2b.JavaScript` | node |
| TypeScript | `e2b.TypeScript` | ts-node |
| Bash | `e2b.Bash` | bash |
| Go | `e2b.GoLang` | go |
| Rust | `e2b.Rust` | rust |
| Java | `e2b.Java` | java |

### Execution 结构体

```go
type Execution struct {
    Results        []*Result      // 执行结果
    Logs           Logs           // 日志输出
    Error          *ExecutionError // 错误信息
    ExecutionCount int            // 执行次数
}

type Result struct {
    Text         string // 文本表示
    HTML         string // HTML 表示
    Markdown     string // Markdown 表示
    SVG          string // SVG 表示
    JSON         string // JSON 表示
    IsMainResult bool   // 是否为主要结果
    Chart        any    // 图表数据
}

type Log struct {
    Line      string // 日志行
    Timestamp int64  // 时间戳
    IsError   bool   // 是否为错误输出
}
```

### 文件系统 API

```go
// 读取文件
content, err := sbx.Files.Read(path string) (string, error)
data, err := sbx.Files.ReadBytes(path string) ([]byte, error)

// 写入文件
err := sbx.Files.Write(path string, data []byte) error
err := sbx.Files.WriteString(path string, content string) error

// 列出目录
entries, err := sbx.Files.List(path string) ([]*EntryInfo, error)

// 创建目录
err := sbx.Files.MakeDir(path string) error

// 检查是否存在
exists, err := sbx.Files.Exists(path string) (bool, error)

// 获取信息
info, err := sbx.Files.GetInfo(path string) (*EntryInfo, error)

// 删除文件/目录
err := sbx.Files.Remove(path string) error
```

### EntryInfo 结构体

```go
type EntryInfo struct {
    Name         string     // 名称
    Type         FileType   // 类型 ("file" 或 "dir")
    Path         string     // 路径
    Size         int64      // 大小（字节）
    Mode         int        // 文件模式
    Permissions  string     // 权限字符串
    Owner        string     // 所有者
    Group        string     // 组
    ModifiedTime *time.Time // 修改时间
}
```

## 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `E2B_API_KEY` | API 密钥 | - |
| `E2B_API_URL` | API 地址 | `https://cn-yangzhou-1-sandbox.qiniuapi.com` |
| `E2B_LOCAL_MODE` | 本地模式 | `false` |

## 示例

查看 `examples/` 目录获取更多示例：

- `examples/main.go` - 基础用法示例（Python 和 JavaScript 执行）
- `examples/filesystem/main.go` - 文件系统操作示例

运行示例：

```bash
# 设置 API Key
export E2B_API_KEY=your-api-key

# 运行基础示例
go run examples/main.go

# 运行文件系统示例
go run examples/filesystem/main.go
```

## 错误处理

```go
execution, err := sbx.RunCode(code, &e2b.RunCodeOpts{
    Language: e2b.Python,
})
if err != nil {
    log.Fatalf("执行失败: %v", err)
}

// 检查执行错误
if execution.Error != nil {
    fmt.Printf("错误名称: %s\n", execution.Error.Name)
    fmt.Printf("错误值: %s\n", execution.Error.Value)
    fmt.Printf("堆栈: %s\n", execution.Error.Traceback)
}
```

## 图表支持

SDK 支持解析代码执行返回的图表数据：

```go
type Chart struct {
    Type     ChartType // "line", "scatter", "bar", "pie", "box_and_whisker", "superchart"
    Title    string
    Elements []any
}

// 支持的图表类型
type LineChart struct { ... }
type ScatterChart struct { ... }
type BarChart struct { ... }
type PieChart struct { ... }
type BoxAndWhiskerChart struct { ... }
type SuperChart struct { ... }
```

## 实现说明

### 文件系统 API

本 SDK 的文件系统操作通过以下方式实现：

| 操作 | 实现方式 | 状态 |
|------|----------|------|
| 读取文件 | REST API (`GET /files`) | ✅ |
| 写入文件 | REST API (`POST /files`, multipart/form-data) | ✅ |
| 列出目录 | 代码执行 (Python) | ✅ |
| 创建目录 | 代码执行 (Python) | ✅ |
| 删除文件/目录 | 代码执行 (Python) | ✅ |
| 检查存在 | 代码执行 (Python) | ✅ |
| 获取信息 | 代码执行 (Python) | ✅ |

> **注意**: E2B 官方 SDK 使用 gRPC/Connect-RPC 协议进行大部分文件系统操作。本 SDK 使用代码执行作为临时方案，未来计划支持 gRPC。

### 端口说明

| 端口 | 用途 |
|------|------|
| 49999 (JupyterPort) | 代码执行 |
| 49983 (EnvdPort) | 文件系统 REST API |

## 许可证

MIT License

## 相关链接

- [Qiniu Cloud](https://www.qiniu.com/)
- [E2B Documentation](https://e2b.dev/docs)
