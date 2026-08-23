/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : OpenCode JSON/JSON5 provider 配置写入
 * @File          : OpenCode 客户端适配器
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package clientconfig

import (
	"fmt"
	"strings"

	"codexrelay/internal/config"
)

func configureOpenCode(path, endpoint, key string, models []config.ModelEntry, defaultModel string) error {
	if err := backupClientFile(path); err != nil {
		return err
	}
	value, err := readJSONObject(path)
	if err != nil {
		return fmt.Errorf("解析 OpenCode 配置失败，请手动配置: %w", err)
	}
	providers, ok := value["provider"].(map[string]any)
	if !ok {
		providers = map[string]any{}
		value["provider"] = providers
	}
	provider := map[string]any{
		"npm":     "@ai-sdk/openai-compatible",
		"name":    "CodexRelay",
		"options": map[string]any{"baseURL": endpoint, "apiKey": key},
	}
	if len(models) > 0 {
		modelMap := map[string]any{}
		for _, model := range models {
			name := strings.TrimSpace(model.Name)
			if name == "" {
				name = model.ID
			}
			modelMap[model.ID] = map[string]any{"name": name}
		}
		provider["models"] = modelMap
	}
	providers["codexrelay"] = provider
	return writeJSONObject(path, value)
}
