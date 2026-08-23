/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : 本地访问认证与安全错误输出
 */
package relay

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
)

func validLocalToken(request *http.Request, expected string) bool {
	if expected == "" {
		return false
	}
	provided := strings.TrimSpace(request.Header.Get("Authorization"))
	if scheme, token, ok := strings.Cut(provided, " "); ok && strings.EqualFold(scheme, "Bearer") {
		provided = strings.TrimSpace(token)
	}
	if provided == "" {
		provided = strings.TrimSpace(request.Header.Get("X-API-Key"))
	}
	return len(provided) == len(expected) && subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

func writeProxyError(writer http.ResponseWriter, status int, message string) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(map[string]any{
		"error": map[string]string{"message": message, "type": "codex_relay_error"},
	})
}

func SanitizeError(err error) string {
	if err == nil {
		return ""
	}
	text := err.Error()
	if len(text) > 300 {
		text = text[:300]
	}
	return text
}
