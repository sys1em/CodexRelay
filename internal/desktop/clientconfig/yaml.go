/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : 客户端 YAML 配置的行级更新辅助
 * @File          : YAML 配置辅助
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package clientconfig

import "strings"

func upsertYAMLLine(lines []string, key, replacement string) []string {
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), key+":") {
			lines[i] = replacement
			return lines
		}
	}
	return append(lines, replacement)
}
