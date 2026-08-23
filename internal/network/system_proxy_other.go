//go:build !windows

/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : 非 Windows 系统代理检测
 */
package network

import (
	"net/http"
	"net/url"
)

func DetectSystemProxy() SystemProxyInfo {
	return SystemProxyInfo{Source: "环境变量", Note: "使用 HTTP_PROXY / HTTPS_PROXY / NO_PROXY"}
}

func systemProxyFunc(_ SystemProxyInfo) func(*http.Request) (*url.URL, error) {
	return http.ProxyFromEnvironment
}
