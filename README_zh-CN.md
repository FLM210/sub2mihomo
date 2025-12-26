# 订阅链接转Mihomo配置转换器

一个通过HTTP服务将订阅链接转换为mihomo配置文件的Go应用程序。

## 功能

- 转换各种代理订阅格式（Shadowsocks、V2Ray/Vmess、Trojan、VLESS、ShadowsocksR、Trojan-go）
- 提供JSON和YAML输出格式（YAML是mihomo的标准格式）
- 简单的HTTP API，便于与其他工具集成
- Web界面用于手动转换
- 支持带有查询参数和标签的代理URL

## 项目结构

```
sub2mihomo/
├── main.go                 # 主应用程序入口点
├── internal/
│   ├── models/             # 数据模型和结构
│   │   └── config.go       # 配置结构
│   ├── handlers/           # HTTP请求处理器
│   │   └── convert_handler.go # 转换端点处理器
│   ├── parsers/            # 代理URL解析逻辑
│   │   └── proxy_parser.go # 代理解析函数
│   └── utils/              # 工具函数
│       └── http_utils.go   # HTTP工具函数
├── README.md
└── go.mod
```

## 安装

1. 确保您的系统上安装了Go
2. 克隆或下载此仓库
3. 运行应用程序：

```bash
go run main.go
```

或者构建并运行：

```bash
go build -o sub2mihomo main.go
./sub2mihomo
```

## 使用方法

应用程序在 `http://localhost:8080` 上启动一个Web服务器，包含以下端点：

### Web界面
- `GET /` - 一个简单的Web界面，用于粘贴订阅URL并进行转换

### API端点
- `POST /convert` - 将订阅URL转换为mihomo配置

#### API使用示例

**JSON请求：**
```bash
curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -d '{"url":"your_subscription_url_here"}'
```

**表单请求：**
```bash
curl -X POST http://localhost:8080/convert \
  -d "url=your_subscription_url_here"
```

**获取YAML输出：**
```bash
curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -H "Accept: application/yaml" \
  -d '{"url":"your_subscription_url_here"}'
```

## 支持的代理类型

- Shadowsocks (ss://)
- V2Ray/Vmess (vmess://)
- Trojan (trojan://)
- VLESS (vless://)
- ShadowsocksR (ssr://)
- Trojan-go (trojan-go://)

## 配置输出

应用程序生成与mihomo兼容的配置，包括：
- 从订阅中解析的代理
- 默认的"PROXY"代理组
- 用于常见用例的基本规则
- mihomo的一般设置

## 示例输出

输出包括：
- `proxies`：从订阅解析的代理配置列表
- `proxy-groups`：带有默认"PROXY"组的代理组
- `rules`：用于路由流量的基本规则集
- `general`：mihomo的一般设置

## 许可证

MIT许可证