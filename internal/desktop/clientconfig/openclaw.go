/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : OpenClaw JSON/JSON5 provider 和默认模型配置写入
 * @File          : OpenClaw 客户端适配器
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package clientconfig

import (
	"fmt"
	"strings"

	"codexrelay/internal/config"
)

func configureOpenClaw(path, endpoint, key string, catalog []config.ModelEntry, defaultModel string) error {
	if err := backupClientFile(path); err != nil {
		return err
	}
	value, err := readJSONObject(path)
	if err != nil {
		return fmt.Errorf("解析 OpenClaw JSON/JSON5 配置失败，请手动配置: %w", err)
	}
	modelRoot, ok := value["models"].(map[string]any)
	if !ok {
		modelRoot = map[string]any{}
		value["models"] = modelRoot
	}
	providers, ok := modelRoot["providers"].(map[string]any)
	if !ok {
		providers = map[string]any{}
		modelRoot["providers"] = providers
	}
	provider := map[string]any{"baseUrl": endpoint, "api": "openai-completions", "apiKey": key}
	if len(catalog) > 0 {
		modelEntries := make([]any, 0, len(catalog))
		for _, model := range catalog {
			name := strings.TrimSpace(model.Name)
			if name == "" {
				name = model.ID
			}
			entry := map[string]any{"id": model.ID, "name": name}
			if model.ContextWindow > 0 {
				entry["contextWindow"] = model.ContextWindow
			}
			modelEntries = append(modelEntries, entry)
		}
		provider["models"] = modelEntries
	}
	providers["codexrelay"] = provider
	if len(catalog) > 0 {
		if defaultModel == "" || !containsModel(catalog, defaultModel) {
			defaultModel = catalog[0].ID
		}
		agents, ok := value["agents"].(map[string]any)
		if !ok {
			agents = map[string]any{}
			value["agents"] = agents
		}
		defaults, ok := agents["defaults"].(map[string]any)
		if !ok {
			defaults = map[string]any{}
			agents["defaults"] = defaults
		}
		modelConfig, ok := defaults["model"].(map[string]any)
		if !ok {
			modelConfig = map[string]any{}
			defaults["model"] = modelConfig
		}
		modelConfig["primary"] = "codexrelay/" + defaultModel
		allowedModels, ok := defaults["models"].(map[string]any)
		if !ok {
			allowedModels = map[string]any{}
			defaults["models"] = allowedModels
		}
		for _, model := range catalog {
			allowedModels["codexrelay/"+model.ID] = map[string]any{}
		}
	}
	return writeJSONObject(path, value)
}
