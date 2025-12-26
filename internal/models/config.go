package models

// SubscriptionRequest represents the request body for subscription conversion
type SubscriptionRequest struct {
	URL    string   `json:"url"`
	Filter []string `json:"filter,omitempty"`
}

// MihomoConfig represents the basic structure of a mihomo config
type MihomoConfig struct {
	ProxyProviders map[string]interface{} `json:"proxy-providers,omitempty" yaml:"proxy-providers,omitempty"`
	Proxies        []interface{}          `json:"proxies,omitempty" yaml:"proxy-providers,omitempty"`
	ProxyGroups    []interface{}          `json:"proxy-groups,omitempty" yaml:"proxy-groups,omitempty"`
	RuleProviders  map[string]interface{} `json:"rule-providers,omitempty" yaml:"rule-providers,omitempty"`
	Rules          []interface{}          `json:"rules,omitempty" yaml:"rules,omitempty"`
	General        map[string]interface{} `json:"general,omitempty" yaml:"general,omitempty"`
}