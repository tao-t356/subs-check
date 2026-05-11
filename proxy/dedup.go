package proxies

import (
	"fmt"
	"sort"
	"strings"
)

func DeduplicateProxies(proxies []map[string]any) []map[string]any {
	seenKeys := make(map[string]bool)
	result := make([]map[string]any, 0, len(proxies))

	for _, proxy := range proxies {
		server, _ := proxy["server"].(string)
		if server == "" {
			continue
		}
		key := proxyIdentityKey(proxy)

		if !seenKeys[key] {
			seenKeys[key] = true
			result = append(result, proxy)
		}
	}

	return result
}

func proxyIdentityKey(proxy map[string]any) string {
	t := strings.ToLower(stringValue(proxy["type"]))
	fields := commonIdentityFields()
	fields = append(fields, protocolIdentityFields(t)...)
	sort.Strings(fields)

	seen := make(map[string]struct{}, len(fields))
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		if value, ok := proxy[field]; ok {
			parts = append(parts, field+"="+stableValue(value))
		}
	}

	// If we don't know the protocol-specific identity well enough, fall back to
	// a stable serialization of all non-metadata fields. This avoids collapsing
	// distinct nodes that share only server/port/password but differ in transport
	// details introduced by newer Mihomo protocols.
	if len(parts) <= 3 {
		keys := make([]string, 0, len(proxy))
		for k := range proxy {
			if k == "name" || k == "sub_url" || k == "sub_tag" {
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts = parts[:0]
		for _, k := range keys {
			parts = append(parts, k+"="+stableValue(proxy[k]))
		}
	}

	return strings.Join(parts, "|")
}

func commonIdentityFields() []string {
	return []string{
		"type", "server", "port", "username", "password", "uuid", "cipher",
		"network", "tls", "sni", "servername", "alpn", "client-fingerprint",
		"fingerprint", "skip-cert-verify", "udp", "dialer-proxy",
	}
}

func protocolIdentityFields(t string) []string {
	switch t {
	case "ss", "ssr":
		return []string{"plugin", "plugin-opts", "obfs", "protocol", "protocol-param"}
	case "vmess":
		return []string{"alterId", "aid", "ws-opts", "grpc-opts", "h2-opts", "http-opts"}
	case "vless":
		return []string{"flow", "reality-opts", "ws-opts", "grpc-opts", "h2-opts", "http-opts"}
	case "trojan":
		return []string{"ws-opts", "grpc-opts", "reality-opts"}
	case "hysteria", "hysteria2", "hy2":
		return []string{"auth", "auth-str", "obfs", "obfs-password", "obfs_password", "up", "down", "ports", "recv-window-conn", "recv-window"}
	case "tuic":
		return []string{"congestion-controller", "udp-relay-mode", "disable-sni", "reduce-rtt"}
	case "wireguard":
		return []string{"private-key", "public-key", "peer-public-key", "pre-shared-key", "ip", "ipv6", "allowed-ips", "reserved", "mtu"}
	default:
		return nil
	}
}

func stringValue(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

func stableValue(v any) string {
	switch x := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, k+":"+stableValue(x[k]))
		}
		return "{" + strings.Join(parts, ",") + "}"
	case map[any]any:
		keys := make([]string, 0, len(x))
		values := make(map[string]any, len(x))
		for k, v := range x {
			ks := fmt.Sprint(k)
			keys = append(keys, ks)
			values[ks] = v
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, k+":"+stableValue(values[k]))
		}
		return "{" + strings.Join(parts, ",") + "}"
	case []any:
		parts := make([]string, 0, len(x))
		for _, item := range x {
			parts = append(parts, stableValue(item))
		}
		return "[" + strings.Join(parts, ",") + "]"
	case []string:
		return "[" + strings.Join(x, ",") + "]"
	default:
		return fmt.Sprint(v)
	}
}
