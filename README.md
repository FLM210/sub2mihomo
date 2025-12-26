# Subscription to Mihomo Converter

A Go application that converts subscription links to mihomo configuration files via an HTTP service.

## Features

- Converts various proxy subscription formats (Shadowsocks, V2Ray/Vmess, Trojan, VLESS, ShadowsocksR, Trojan-go)
- Provides both JSON and YAML output formats (YAML is the standard for mihomo)
- Simple HTTP API for integration with other tools
- Web interface for manual conversion
- Support for proxy URLs with query parameters and tags

## Project Structure

```
sub2mihomo/
├── main.go                 # Main application entry point
├── internal/
│   ├── models/             # Data models and structures
│   │   └── config.go       # Configuration structures
│   ├── handlers/           # HTTP request handlers
│   │   └── convert_handler.go # Conversion endpoint handlers
│   ├── parsers/            # Proxy URL parsing logic
│   │   └── proxy_parser.go # Proxy parsing functions
│   └── utils/              # Utility functions
│       └── http_utils.go   # HTTP utility functions
├── README.md
└── go.mod
```

## Installation

1. Make sure you have Go installed on your system
2. Clone or download this repository
3. Run the application:

```bash
go run main.go
```

Or build and run:

```bash
go build -o sub2mihomo main.go
./sub2mihomo```

## Usage

The application starts a web server on `http://localhost:8080` with the following endpoints:

### Web Interface
- `GET /` - A simple web interface to paste subscription URLs and convert them

### API Endpoints
- `POST /convert` - Convert subscription URL to mihomo config

#### API Usage Examples

**JSON Request:**
```bash
curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -d '{"url":"your_subscription_url_here"}'
```

**Form Request:**
```bash
curl -X POST http://localhost:8080/convert \
  -d "url=your_subscription_url_here"
```

**Get YAML Output:**
```bash
curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -H "Accept: application/yaml" \
  -d '{"url":"your_subscription_url_here"}'
```

## Supported Proxy Types

- Shadowsocks (ss://)
- V2Ray/Vmess (vmess://)
- Trojan (trojan://)
- VLESS (vless://)
- ShadowsocksR (ssr://)
- Trojan-go (trojan-go://)

## Configuration Output

The application generates a mihomo-compatible configuration with:
- Proxies parsed from the subscription
- A default "PROXY" proxy group
- Basic rules for common use cases
- General settings for mihomo

## Example Output

The output includes:
- `proxies`: List of proxy configurations parsed from the subscription
- `proxy-groups`: Proxy groups with a default "PROXY" group
- `rules`: Basic rule set for routing traffic
- `general`: General mihomo settings

## License

MIT License