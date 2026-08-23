//go:build windows

/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : Windows 系统代理解析测试
 */
package network

import "testing"

func TestParseWindowsProxyServer(t *testing.T) {
	httpProxy, httpsProxy := parseWindowsProxyServer("http=127.0.0.1:7890;https=127.0.0.1:7891")
	if httpProxy != "http://127.0.0.1:7890" || httpsProxy != "http://127.0.0.1:7891" {
		t.Fatalf("unexpected proxies: %q %q", httpProxy, httpsProxy)
	}
}

func TestWindowsProxyBypass(t *testing.T) {
	for _, host := range []string{"localhost", "127.0.0.1", "printer", "api.example.com"} {
		if !shouldBypassWindowsProxy(host, "<local>;*.example.com") {
			t.Errorf("host %q should bypass", host)
		}
	}
	if shouldBypassWindowsProxy("api.other.com", "<local>;*.example.com") {
		t.Fatal("unmatched host should not bypass")
	}
}
