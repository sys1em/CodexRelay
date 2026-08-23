/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : 客户端 TOML 配置的片段级更新辅助
 * @File          : TOML 配置辅助
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package clientconfig

import (
	"strconv"
	"strings"
)

func upsertTomlProvider(raw, providerID, endpoint string) string {
	return upsertTomlProviderWithModel(raw, providerID, endpoint, "")
}

func upsertTomlProviderWithModel(raw, providerID, endpoint, defaultModel string) string {
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	foundTop := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "model_provider") && strings.Contains(line, "=") {
			lines[i] = "model_provider = \"" + providerID + "\""
			foundTop = true
			break
		}
	}
	if !foundTop {
		lines = append([]string{"model_provider = \"" + providerID + "\""}, lines...)
	}
	if strings.TrimSpace(defaultModel) != "" {
		lines = upsertTomlTopLevelLine(lines, "model", strconv.Quote(defaultModel))
	}
	header := "[model_providers." + providerID + "]"
	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == header {
			start = i
			break
		}
	}
	if start < 0 {
		block := []string{"", header, "name = \"CodexRelay\"", "base_url = \"" + endpoint + "\"", "wire_api = \"responses\"", "requires_openai_auth = true"}
		lines = append(lines, block...)
	} else {
		end := len(lines)
		for i := start + 1; i < len(lines); i++ {
			if strings.HasPrefix(strings.TrimSpace(lines[i]), "[") {
				end = i
				break
			}
		}
		section := lines[start+1 : end]
		section = upsertTomlLine(section, "name", "\"CodexRelay\"")
		section = upsertTomlLine(section, "base_url", "\""+endpoint+"\"")
		section = upsertTomlLine(section, "wire_api", "\"responses\"")
		section = upsertTomlLine(section, "requires_openai_auth", "true")
		lines = append(append(append([]string{}, lines[:start+1]...), section...), lines[end:]...)
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n") + "\n"
}

func upsertTomlTopLevelLine(lines []string, key, value string) []string {
	for i, line := range lines {
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, key+" ") || strings.HasPrefix(trimmed, key+"=") {
				lines[i] = key + " = " + value
				return lines
			}
		}
	}
	return append([]string{key + " = " + value}, lines...)
}

func upsertTomlLine(lines []string, key, value string) []string {
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, key+" ") || strings.HasPrefix(trimmed, key+"=") {
			lines[i] = key + " = " + value
			return lines
		}
	}
	return append(lines, key+" = "+value)
}

func upsertTomlSectionValue(lines []string, sectionName, key, value string) []string {
	header := "[" + sectionName + "]"
	section := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == header {
			section = i
			break
		}
	}
	if section < 0 {
		return append(lines, "", header, key+" = "+value)
	}
	end := len(lines)
	for i := section + 1; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "[") {
			end = i
			break
		}
	}
	block := upsertTomlLine(lines[section+1:end], key, value)
	return append(append(append([]string{}, lines[:section+1]...), block...), lines[end:]...)
}

func upsertTomlSection(lines []string, sectionName string, values []string) []string {
	header := "[" + sectionName + "]"
	section := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == header {
			section = i
			break
		}
	}
	if section < 0 {
		return append(append(lines, ""), append([]string{header}, values...)...)
	}
	end := len(lines)
	for i := section + 1; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "[") {
			end = i
			break
		}
	}
	block := append([]string(nil), lines[section+1:end]...)
	for _, value := range values {
		parts := strings.SplitN(value, "=", 2)
		if len(parts) == 2 {
			block = upsertTomlLine(block, strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
	return append(append(append([]string{}, lines[:section+1]...), block...), lines[end:]...)
}
