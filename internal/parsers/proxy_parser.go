package parsers

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// parseProxy parses a proxy URL and converts it to mihomo proxy config
func ParseProxy(proxyURL string) (map[string]interface{}, error) {
	// This is a simplified parser - in a real implementation you'd want to handle
	// different proxy types (vmess, vless, trojan, ss, etc.) properly
	if strings.HasPrefix(proxyURL, "ss://") {
		return parseShadowsocks(proxyURL)
	} else if strings.HasPrefix(proxyURL, "vmess://") {
		return parseVmess(proxyURL)
	} else if strings.HasPrefix(proxyURL, "trojan://") {
		return parseTrojan(proxyURL)
	} else if strings.HasPrefix(proxyURL, "vless://") {
		return parseVless(proxyURL)
	} else if strings.HasPrefix(proxyURL, "trojan-go://") {
		return parseTrojanGo(proxyURL)
	} else if strings.HasPrefix(proxyURL, "ssr://") {
		return parseShadowsocksR(proxyURL)
	}

	return nil, fmt.Errorf("unsupported proxy type: %s", proxyURL)
}

// parseShadowsocks parses a shadowsocks URL
func parseShadowsocks(ssURL string) (map[string]interface{}, error) {
	// Remove the ss:// prefix
	urlStr := strings.TrimPrefix(ssURL, "ss://")

	// Handle both old and new format
	var method, password, server, port string
	var tag string

	if idx := strings.Index(urlStr, "#"); idx != -1 {
		// Extract tag name and URL decode it
		tag = urlStr[idx+1:]

		// Check if the tag contains URL encoded characters
		if strings.Contains(tag, "%") {
			// URL decode the tag
			decodedTag, err := url.QueryUnescape(tag)
			if err != nil {
				// Fallback to manual replacements if QueryUnescape fails
				decodedTag = strings.ReplaceAll(tag, "%20", " ")
				decodedTag = strings.ReplaceAll(decodedTag, "%F0%9F%87%B9%F0%9F%87%BC", "🇹🇼") // Taiwan flag emoji
			}
			tag = decodedTag
		}

		// Remove emoji characters from the tag
		tag = removeEmojis(tag)

		urlStr = urlStr[:idx]
	}

	if idx := strings.Index(urlStr, "@"); idx != -1 {
		// New format: ss://method:password@server:port OR ss://base64(method:password)@server:port
		authAndServer := urlStr[:idx]
		serverAndPort := urlStr[idx+1:]

		// Try to decode the authentication part as base64
		// Try different padding lengths since it might be missing
		var decodedAuth []byte
		var isBase64Encoded = false

		for i := 0; i <= 3; i++ {
			paddedAuth := authAndServer
			for j := 0; j < i; j++ {
				paddedAuth += "="
			}

			decoded, err := base64.StdEncoding.DecodeString(paddedAuth)
			if err == nil {
				decodedAuth = decoded
				isBase64Encoded = true
				break
			}
		}

		if isBase64Encoded {
			// It's base64 encoded
			authStr := string(decodedAuth)
			parts := strings.Split(authStr, ":")
			if len(parts) >= 2 {
				method = parts[0]
				// Combine remaining parts as password (in case password itself contains colons)
				password = strings.Join(parts[1:], ":")
			}
		} else {
			// Not base64 encoded, use regular format
			if idx2 := strings.Index(authAndServer, ":"); idx2 != -1 {
				method = authAndServer[:idx2]
				password = authAndServer[idx2+1:]
			}
		}

		if idx2 := strings.Index(serverAndPort, ":"); idx2 != -1 {
			server = serverAndPort[:idx2]
			port = serverAndPort[idx2+1:]
		}
	} else {
		// Old format: ss://base64(method:password)@server:port
		if idx := strings.Index(urlStr, "@"); idx != -1 {
			encoded := urlStr[:idx]
			serverAndPort := urlStr[idx+1:]

			// Try different padding lengths
			var decoded []byte
			var err error
			for i := 0; i <= 3; i++ {
				padded := encoded
				for j := 0; j < i; j++ {
					padded += "="
				}
				decoded, err = base64.StdEncoding.DecodeString(padded)
				if err == nil {
					break
				}
			}

			if err != nil {
				return nil, err
			}

			authStr := string(decoded)
			if idx2 := strings.Index(authStr, ":"); idx2 != -1 {
				method = authStr[:idx2]
				password = authStr[idx2+1:]
			}

			if idx2 := strings.Index(serverAndPort, ":"); idx2 != -1 {
				server = serverAndPort[:idx2]
				port = serverAndPort[idx2+1:]
			}
		}
	}

	if tag == "" {
		tag = fmt.Sprintf("%s:%s", server, port)
	}

	return map[string]interface{}{
		"name":     tag,
		"type":     "ss",
		"server":   server,
		"port":     port,
		"cipher":   method,
		"password": password,
	}, nil
}

// parseVmess parses a vmess URL
func parseVmess(vmessURL string) (map[string]interface{}, error) {
	// Remove the vmess:// prefix
	urlStr := strings.TrimPrefix(vmessURL, "vmess://")

	// Extract tag if present
	var tag string
	if idx := strings.Index(urlStr, "#"); idx != -1 {
		tag = urlStr[idx+1:]

		// Check if the tag contains URL encoded characters
		if strings.Contains(tag, "%") {
			// URL decode the tag
			decodedTag, err := url.QueryUnescape(tag)
			if err != nil {
				// Fallback to manual replacements if QueryUnescape fails
				decodedTag = strings.ReplaceAll(tag, "%20", " ")
				decodedTag = strings.ReplaceAll(decodedTag, "%F0%9F%87%B9%F0%9F%87%BC", "🇹🇼") // Taiwan flag emoji
			}
			tag = decodedTag
		}

		// Remove emoji characters from the tag
		tag = removeEmojis(tag)

		urlStr = urlStr[:idx]
	}

	// Decode the base64 encoded part
	decoded, err := base64.StdEncoding.DecodeString(urlStr)
	if err != nil {
		return nil, err
	}

	// Parse the JSON
	var vmessData map[string]interface{}
	if err := json.Unmarshal(decoded, &vmessData); err != nil {
		return nil, err
	}

	// Extract server and port
	server, _ := vmessData["add"].(string)
	portStr, _ := vmessData["port"].(string)
	port, err := strconv.Atoi(portStr)
	if err != nil {
		// If port is not a string, try as number
		if portNum, ok := vmessData["port"].(float64); ok {
			port = int(portNum)
		} else {
			port = 443 // default port
		}
	}

	// Map vmess fields to mihomo format
	proxy := map[string]interface{}{
		"name":   tag,
		"type":   "vmess",
		"server": server,
		"port":   port,
		"uuid":   vmessData["id"],
	}

	if aid, ok := vmessData["aid"]; ok {
		if aidNum, ok := aid.(float64); ok {
			proxy["alterId"] = int(aidNum)
		} else if aidNum, ok := aid.(int); ok {
			proxy["alterId"] = aidNum
		} else {
			proxy["alterId"] = 0 // default value
		}
	} else {
		proxy["alterId"] = 0
	}

	if tag == "" {
		proxy["name"] = fmt.Sprintf("%s:%d", server, port)
	}

	if security, ok := vmessData["scy"]; ok {
		proxy["cipher"] = security
	} else {
		proxy["cipher"] = "auto"
	}

	if tls, ok := vmessData["tls"]; ok && tls == "tls" {
		proxy["tls"] = true
	}

	if sni, ok := vmessData["sni"]; ok {
		proxy["servername"] = sni
	}

	if net, ok := vmessData["net"]; ok {
		proxy["network"] = net
		switch net {
		case "ws":
			wsOpts := make(map[string]interface{})
			if host, ok := vmessData["host"]; ok {
				headers := make(map[string]string)
				headers["Host"] = fmt.Sprintf("%v", host)
				wsOpts["headers"] = headers
			}
			if path, ok := vmessData["path"]; ok {
				wsOpts["path"] = path
			}
			if len(wsOpts) > 0 {
				proxy["ws-opts"] = wsOpts
			}
		case "grpc":
			if path, ok := vmessData["path"]; ok {
				proxy["grpc-opts"] = map[string]interface{}{
					"grpc-service-name": path,
				}
			}
		}
	}

	return proxy, nil
}

// parseTrojan parses a trojan URL
func parseTrojan(trojanURL string) (map[string]interface{}, error) {
	// Remove the trojan:// prefix
	urlStr := strings.TrimPrefix(trojanURL, "trojan://")

	// Extract tag if present
	var tag string
	if idx := strings.Index(urlStr, "#"); idx != -1 {
		tag = urlStr[idx+1:]

		// Check if the tag contains URL encoded characters
		if strings.Contains(tag, "%") {
			// URL decode the tag
			decodedTag, err := url.QueryUnescape(tag)
			if err != nil {
				// Fallback to manual replacements if QueryUnescape fails
				decodedTag = strings.ReplaceAll(tag, "%20", " ")
				decodedTag = strings.ReplaceAll(decodedTag, "%F0%9F%87%B9%F0%9F%87%BC", "🇹🇼") // Taiwan flag emoji
			}
			tag = decodedTag
		}

		// Remove emoji characters from the tag
		tag = removeEmojis(tag)

		urlStr = urlStr[:idx]
	}
	var password, server, port string
	var query string

	if idx := strings.Index(urlStr, "@"); idx != -1 {
		password = urlStr[:idx]
		rest := urlStr[idx+1:]

		// Check for query parameters
		if idx2 := strings.Index(rest, "?"); idx2 != -1 {
			query = rest[idx2+1:]
			rest = rest[:idx2]
		}

		if idx2 := strings.LastIndex(rest, ":"); idx2 != -1 {
			server = rest[:idx2]
			port = rest[idx2+1:]
		}
	}

	if tag == "" {
		tag = fmt.Sprintf("%s:%s", server, port)
	}

	proxy := map[string]interface{}{
		"name":             tag,
		"type":             "trojan",
		"server":           server,
		"port":             port,
		"password":         password,
		"sni":              server,
		"skip-cert-verify": true,
	}

	// Parse query parameters
	params := parseQuery(query)
	if sni, exists := params["sni"]; exists {
		proxy["sni"] = sni
	}
	if alpn, exists := params["alpn"]; exists {
		proxy["alpn"] = alpn
	}
	if fp, exists := params["fp"]; exists {
		proxy["client-fingerprint"] = fp
	}
	if allowInsecure, exists := params["allowInsecure"]; exists {
		if allowInsecure == "1" || allowInsecure == "true" {
			proxy["skip-cert-verify"] = true
		} else {
			proxy["skip-cert-verify"] = false
		}
	}

	return proxy, nil
}

// parseVless parses a vless URL
func parseVless(vlessURL string) (map[string]interface{}, error) {
	// Remove the vless:// prefix
	urlStr := strings.TrimPrefix(vlessURL, "vless://")

	// Extract tag if present
	var tag string
	if idx := strings.Index(urlStr, "#"); idx != -1 {
		tag = urlStr[idx+1:]

		// Check if the tag contains URL encoded characters
		if strings.Contains(tag, "%") {
			// URL decode the tag
			decodedTag, err := url.QueryUnescape(tag)
			if err != nil {
				// Fallback to manual replacements if QueryUnescape fails
				decodedTag = strings.ReplaceAll(tag, "%20", " ")
				decodedTag = strings.ReplaceAll(decodedTag, "%F0%9F%87%B9%F0%9F%87%BC", "🇹🇼") // Taiwan flag emoji
			}
			tag = decodedTag
		}

		// Remove emoji characters from the tag
		tag = removeEmojis(tag)

		urlStr = urlStr[:idx]
	}

	// Parse the URL
	var uuid, server, port string
	var query string

	if idx := strings.Index(urlStr, "@"); idx != -1 {
		uuid = urlStr[:idx]
		rest := urlStr[idx+1:]

		// Check for query parameters
		if idx2 := strings.Index(rest, "?"); idx2 != -1 {
			query = rest[idx2+1:]
			rest = rest[:idx2]
		}

		if idx2 := strings.LastIndex(rest, ":"); idx2 != -1 {
			server = rest[:idx2]
			port = rest[idx2+1:]
		}
	}

	if tag == "" {
		tag = fmt.Sprintf("%s:%s", server, port)
	}

	proxy := map[string]interface{}{
		"name":             tag,
		"type":             "vless",
		"server":           server,
		"port":             port,
		"uuid":             uuid,
		"skip-cert-verify": true,
	}

	// Parse query parameters
	params := parseQuery(query)
	if flow, exists := params["flow"]; exists {
		proxy["flow"] = flow
	}
	if sni, exists := params["sni"]; exists {
		proxy["servername"] = sni
	}
	if typeParam, exists := params["type"]; exists {
		proxy["network"] = typeParam
		switch typeParam {
		case "ws":
			wsOpts := make(map[string]interface{})
			if host, exists := params["host"]; exists {
				headers := make(map[string]string)
				headers["Host"] = host
				wsOpts["headers"] = headers
			}
			if path, exists := params["path"]; exists {
				// Check if the path contains URL encoded characters
				decodedPath := path
				if strings.Contains(path, "%") {
					// URL decode the path
					decodedTemp, err := url.QueryUnescape(path)
					if err != nil {
						// Fallback to manual replacements if QueryUnescape fails
						decodedTemp = strings.ReplaceAll(path, "%2F", "/")        // forward slash
						decodedTemp = strings.ReplaceAll(decodedTemp, "%3F", "?") // question mark
						decodedTemp = strings.ReplaceAll(decodedTemp, "%3D", "=") // equals
						decodedTemp = strings.ReplaceAll(decodedTemp, "%26", "&") // ampersand
					}
					decodedPath = decodedTemp
				}
				wsOpts["path"] = decodedPath
			}
			if len(wsOpts) > 0 {
				proxy["ws-opts"] = wsOpts
			}
		case "grpc":
			if serviceName, exists := params["serviceName"]; exists {
				proxy["grpc-opts"] = map[string]interface{}{
					"grpc-service-name": serviceName,
				}
			}
		}
	}
	if alpn, exists := params["alpn"]; exists {
		proxy["alpn"] = alpn
	}
	if fp, exists := params["fp"]; exists {
		proxy["client-fingerprint"] = fp
	}

	return proxy, nil
}

// parseTrojanGo parses a trojan-go URL
func parseTrojanGo(trojanGoURL string) (map[string]interface{}, error) {
	// For now, treat trojan-go similar to trojan
	return parseTrojan(strings.Replace(trojanGoURL, "trojan-go://", "trojan://", 1))
}

// parseShadowsocksR parses a shadowsocksR URL
func parseShadowsocksR(ssrURL string) (map[string]interface{}, error) {
	// Remove the ssr:// prefix
	urlStr := strings.TrimPrefix(ssrURL, "ssr://")

	// Decode the base64 encoded part
	decoded, err := base64.URLEncoding.DecodeString(strings.TrimRight(urlStr, "="))
	if err != nil {
		// Try standard base64 if URL encoding fails
		decoded, err = base64.StdEncoding.DecodeString(strings.TrimRight(urlStr, "="))
		if err != nil {
			return nil, err
		}
	}

	parts := strings.Split(string(decoded), ":")
	if len(parts) < 6 {
		return nil, fmt.Errorf("invalid ssr format")
	}

	server := parts[0]
	portStr := parts[1]
	protocol := parts[2]
	method := parts[3]
	obfs := parts[4]
	base64Pass := parts[5]

	// Extract additional parameters from query string if present
	var additionalParams string
	if idx := strings.Index(base64Pass, "/?"); idx != -1 {
		additionalParams = base64Pass[idx+2:]
		base64Pass = base64Pass[:idx]
	}

	password, err := base64.URLEncoding.DecodeString(strings.TrimRight(base64Pass, "="))
	if err != nil {
		password, err = base64.StdEncoding.DecodeString(strings.TrimRight(base64Pass, "="))
		if err != nil {
			return nil, err
		}
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	proxy := map[string]interface{}{
		"name":     fmt.Sprintf("%s:%s", server, portStr),
		"type":     "ssr",
		"server":   server,
		"port":     port,
		"cipher":   method,
		"password": string(password),
		"protocol": protocol,
		"obfs":     obfs,
	}

	// Parse additional parameters
	params := parseQuery(additionalParams)
	if params["remarks"] != "" {
		remarks := params["remarks"]

		// Check if the remarks contain URL encoded characters
		if strings.Contains(remarks, "%") {
			// URL decode the remarks
			decodedRemarks, err := url.QueryUnescape(remarks)
			if err != nil {
				// Fallback to manual replacements if QueryUnescape fails
				decodedRemarks = strings.ReplaceAll(remarks, "%20", " ")
				decodedRemarks = strings.ReplaceAll(decodedRemarks, "%F0%9F%87%B9%F0%9F%87%BC", "🇹🇼") // Taiwan flag emoji
			}
			remarks = decodedRemarks
		}

		// Remove emoji characters from the remarks
		remarks = removeEmojis(remarks)

		proxy["name"] = remarks
	}
	if params["obfsparam"] != "" {
		obfsParam, err := base64.URLEncoding.DecodeString(strings.TrimRight(params["obfsparam"], "="))
		if err != nil {
			obfsParam, err = base64.StdEncoding.DecodeString(strings.TrimRight(params["obfsparam"], "="))
			if err == nil {
				proxy["obfs-param"] = string(obfsParam)
			}
		} else {
			proxy["obfs-param"] = string(obfsParam)
		}
	}
	if params["protoparam"] != "" {
		protoParam, err := base64.URLEncoding.DecodeString(strings.TrimRight(params["protoparam"], "="))
		if err != nil {
			protoParam, err = base64.StdEncoding.DecodeString(strings.TrimRight(params["protoparam"], "="))
			if err == nil {
				proxy["protocol-param"] = string(protoParam)
			}
		} else {
			proxy["protocol-param"] = string(protoParam)
		}
	}

	return proxy, nil
}

// parseQuery parses query string parameters
func parseQuery(query string) map[string]string {
	params := make(map[string]string)

	pairs := strings.Split(query, "&")
	for _, pair := range pairs {
		if idx := strings.Index(pair, "="); idx != -1 {
			key := pair[:idx]
			value := pair[idx+1:]
			// URL decode the value
			value = strings.Replace(value, "%3D", "=", -1)
			value = strings.Replace(value, "%26", "&", -1)
			params[key] = value
		}
	}

	return params
}

// ConvertToMihomo converts subscription content to mihomo config with optional node filtering
func ConvertToMihomo(subContent string, filterKeywords ...string) (*map[string]interface{}, error) {
	// Decode base64 if needed (common for subscription links)
	decodedContent := subContent
	if IsBase64(subContent) {
		decoded, err := base64.StdEncoding.DecodeString(subContent)
		if err == nil {
			decodedContent = string(decoded)
		}
	}

	// Parse the subscription content (usually base64 encoded list of proxy URLs)
	scanner := bufio.NewScanner(strings.NewReader(decodedContent))
	var proxies []interface{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse proxy URL and convert to mihomo proxy config
		proxy, err := ParseProxy(line)
		if err != nil {
			continue // Skip invalid lines
		}

		if proxy != nil {
			// Apply filter if filter keywords are provided
			if len(filterKeywords) > 0 {
				proxyName, ok := proxy["name"].(string)
				if ok {
					shouldInclude := false
					for _, keyword := range filterKeywords {
						if strings.Contains(strings.ToLower(proxyName), strings.ToLower(keyword)) {
							shouldInclude = true
							break
						}
					}
					if shouldInclude {
						proxies = append(proxies, proxy)
					}
				}
			} else {
				// No filter, add all proxies
				proxies = append(proxies, proxy)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Create a basic mihomo config
	config := map[string]interface{}{
		"proxies": proxies,
		"proxy-groups": []interface{}{
			map[string]interface{}{
				"name":      "AUTO",
				"type":      "url-test",      // Auto select proxy based on URL test
				"proxies":   []interface{}{}, // Will be filled with all proxies
				"url":       "http://www.gstatic.com/generate_204",
				"interval":  "300",
				"tolerance": "50",
			},
			map[string]interface{}{
				"name":    "PROXY",
				"type":    "select",
				"proxies": []interface{}{"AUTO", "DIRECT"}, // Default to using AUTO group
			},
		},
		"mode":                "rule",
		"log-level":           "info",
		"allow-lan":           true,           // Allow connections from LAN
		"mixed-port":          10801,          // Mixed proxy port for HTTP and SOCKS
		"external-controller": "0.0.0.0:9090", // API controller listening on all interfaces
		"secret":              "sub2mihomo",   // API access secret
		"external-controller-cors": map[string]interface{}{
			"allow-origins":         []string{"*"},
			"allow-private-network": true,
		},
		"rules": []interface{}{
			"MATCH,PROXY",
		},
	}

	// Fill the AUTO group with all available proxies
	if len(proxies) > 0 {
		autoGroup := config["proxy-groups"].([]interface{})[0].(map[string]interface{})
		for _, proxy := range proxies {
			proxyMap := proxy.(map[string]interface{})
			autoGroup["proxies"] = append(autoGroup["proxies"].([]interface{}), proxyMap["name"])
		}
	}

	return &config, nil
}

// IsBase64 checks if a string is base64 encoded
func IsBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// removeEmojis removes emoji characters from a string, keeping only text content
func removeEmojis(str string) string {
	var result strings.Builder
	for _, r := range str {
		// Check if the rune is an emoji or other symbol
		// This is a simplified check - for a more comprehensive solution, use a dedicated library
		if r < 0x1F000 || (r > 0x1F9FF && (r < 0x2600 || r > 0x27BF)) {
			// Not an emoji range, add to result
			result.WriteRune(r)
		}
	}
	return result.String()
}
