package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"sub2mihomo/internal/models"
	"sub2mihomo/internal/parsers"
	"sub2mihomo/internal/utils"

	"gopkg.in/yaml.v3"
)

func ConvertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody models.SubscriptionRequest

	// Try to read as JSON first
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(body, &reqBody); err != nil {
			http.Error(w, "Error parsing JSON", http.StatusBadRequest)
			return
		}
	} else {
		// Handle form data
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		reqBody.URL = r.FormValue("url")

		// Handle filter as comma-separated values
		filterStr := r.FormValue("filter")
		if filterStr != "" {
			reqBody.Filter = strings.Split(filterStr, ",")
			// Trim spaces from each filter value
			for i, f := range reqBody.Filter {
				reqBody.Filter[i] = strings.TrimSpace(f)
			}
		}
	}

	if reqBody.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Fetch the subscription content
	subContent, err := utils.FetchSubscription(reqBody.URL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching subscription: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert subscription to mihomo config
	var config *map[string]interface{}
	if len(reqBody.Filter) > 0 {
		config, err = parsers.ConvertToMihomo(subContent, reqBody.Filter...)
	} else {
		config, err = parsers.ConvertToMihomo(subContent)
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("Error converting to mihomo config: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to YAML (default format)
	yamlBytes, err := yaml.Marshal(config)
	if err != nil {
		http.Error(w, "Error generating YAML config", http.StatusInternalServerError)
		return
	}

	// Set response headers for YAML file download
	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Content-Disposition", "attachment; filename=config.yaml")
	w.Write(yamlBytes)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Subscription to Mihomo Converter</title>
		<style>
			* {
				margin: 0;
				padding: 0;
				box-sizing: border-box;
				font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
			}
			body {
				background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
				min-height: 100vh;
				padding: 20px;
				display: flex;
				flex-direction: column;
				align-items: center;
			}
			.container {
				max-width: 800px;
				width: 100%;
				background: white;
				border-radius: 10px;
				box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
				padding: 30px;
				margin: 20px 0;
			}
			h1 {
				text-align: center;
				color: #2c3e50;
				margin-bottom: 30px;
				font-size: 28px;
			}
			form {
				margin-bottom: 30px;
			}
			.form-group {
				margin-bottom: 20px;
			}
			label {
				display: block;
				margin-bottom: 8px;
				font-weight: 600;
				color: #34495e;
			}
			input[type="text"] {
				width: 100%%;
				padding: 12px 15px;
				border: 2px solid #e0e0e0;
				border-radius: 6px;
				font-size: 16px;
				transition: border-color 0.3s;
			}
			input[type="text"]:focus {
				outline: none;
				border-color: #3498db;
			}
			input[type="submit"] {
				background-color: #3498db;
				color: white;
				padding: 12px 25px;
				border: none;
				border-radius: 6px;
				cursor: pointer;
				font-size: 16px;
				font-weight: 600;
				width: 100%%;
				transition: background-color 0.3s;
			}
			input[type="submit"]:hover {
				background-color: #2980b9;
			}
			.api-info {
				background-color: #f8f9fa;
				padding: 20px;
				border-radius: 6px;
				border-left: 4px solid #3498db;
			}
			.api-info h3 {
				color: #2c3e50;
				margin-bottom: 10px;
			}
			pre {
				background: #2c3e50;
				color: #ecf0f1;
				padding: 15px;
				border-radius: 4px;
				overflow-x: auto;
				font-size: 14px;
				line-height: 1.4;
			}
			.lang-selector {
				text-align: center;
				margin-bottom: 20px;
			}
			.lang-link {
				margin: 0 10px;
				text-decoration: none;
				padding: 8px 16px;
				border: 1px solid #ddd;
				border-radius: 4px;
				background: #f9f9f9;
				color: #2c3e50;
				transition: all 0.3s;
			}
			.lang-link:hover {
				background: #e9e9e9;
				text-decoration: none;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<div class="lang-selector">
				<a href="https://github.com/FLM210/sub2mihomo/blob/main/README.md" class="lang-link">English</a>
				<a href="https://github.com/FLM210/sub2mihomo/blob/main/README.md_zh-CN.md" class="lang-link">中文</a>
			</div>
			
			<h1>Subscription to Mihomo Config Converter</h1>
			<form action="/convert" method="post">
				<div class="form-group">
					<label for="url">Subscription URL:</label>
					<input type="text" id="url" name="url" placeholder="Enter subscription URL">
				</div>
				<div class="form-group">
					<label for="filter">Filter (optional):</label>
					<input type="text" id="filter" name="filter" placeholder="Enter keywords to filter nodes, comma separated (e.g., Japan,台湾,SG)">
				</div>
				<input type="submit" value="Convert">
			</form>
			
			<div class="api-info">
				<h3>API Usage</h3>
				<p>Or use POST /convert with JSON body: {"url": "your_subscription_url", "filter": ["keyword1", "keyword2"]}</p>
				
				<h3>Example cURL command:</h3>
				<pre>curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -d '{"url":"your_subscription_url_here", "filter":["Japan", "台湾"]}'</pre>
			</div>
		</div>
	</body>
	</html>
	`)
}
