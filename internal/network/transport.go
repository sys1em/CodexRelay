/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : 代理上游 HTTP Transport 构建
 */
package network

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func BuildTransport(settings Settings, systemInfo SystemProxyInfo, listenerPort int) (*http.Transport, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ForceAttemptHTTP2 = true
	transport.DisableCompression = true
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 20
	transport.IdleConnTimeout = 90 * time.Second

	var proxyFunc func(*http.Request) (*url.URL, error)
	switch settings.Mode {
	case "direct":
		proxyFunc = nil
	case "manual":
		proxyURL, err := url.Parse(strings.TrimSpace(settings.ProxyURL))
		if err != nil {
			return nil, err
		}
		proxyFunc = http.ProxyURL(proxyURL)
	default:
		proxyFunc = systemProxyFunc(systemInfo)
	}
	if proxyFunc == nil {
		transport.Proxy = nil
		return transport, nil
	}
	transport.Proxy = func(request *http.Request) (*url.URL, error) {
		proxyURL, err := proxyFunc(request)
		if err != nil || proxyURL == nil {
			return proxyURL, err
		}
		if IsListenerURL(proxyURL, listenerPort) {
			return nil, errors.New("检测到上游代理指向 CodexRelay 自身，已阻止代理循环")
		}
		return proxyURL, nil
	}
	return transport, nil
}
