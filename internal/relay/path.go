/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : 上游 URL 路径与查询拼接
 */
package relay

import (
	"strings"

	"codexrelay/internal/config"
)

// RoutePath extracts a category prefix and leaves the remainder unchanged for the upstream.
func RoutePath(path string) (string, string, bool) {
	parts := splitPath(path)
	if len(parts) == 0 || !config.IsCategory(parts[0]) {
		return "", "", false
	}
	if len(parts) == 1 {
		return parts[0], "/", true
	}
	return parts[0], "/" + strings.Join(parts[1:], "/"), true
}

func JoinTargetPath(basePath, requestPath string) string {
	baseParts := splitPath(basePath)
	requestParts := splitPath(requestPath)
	if len(baseParts) == 0 {
		if len(requestParts) == 0 {
			return "/"
		}
		return "/" + strings.Join(requestParts, "/")
	}
	if len(requestParts) == 0 {
		return "/" + strings.Join(baseParts, "/")
	}
	overlap := 0
	limit := min(len(baseParts), len(requestParts))
	for size := limit; size > 0; size-- {
		matched := true
		for i := 0; i < size; i++ {
			if !strings.EqualFold(baseParts[len(baseParts)-size+i], requestParts[i]) {
				matched = false
				break
			}
		}
		if matched {
			overlap = size
			break
		}
	}
	return "/" + strings.Join(append(baseParts, requestParts[overlap:]...), "/")
}

func splitPath(path string) []string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 1 && parts[0] == "" {
		return nil
	}
	return parts
}

func joinQuery(baseQuery, requestQuery string) string {
	if baseQuery == "" {
		return requestQuery
	}
	if requestQuery == "" {
		return baseQuery
	}
	return baseQuery + "&" + requestQuery
}
