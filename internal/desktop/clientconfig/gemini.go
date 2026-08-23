/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Gemini 的 .env 与 settings.json 配置写入
 * @File          : Gemini 客户端适配器
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package clientconfig

import (
	"encoding/json"
	"fmt"
	"os"

	"codexrelay/internal/config"
)

func configureGeminiSettings(path string) error {
	if !pathExists(path) {
		return nil
	}
	if err := backupClientFile(path); err != nil {
		return err
	}
	value := map[string]any{}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取 Gemini settings.json 失败: %w", err)
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("解析 Gemini settings.json 失败，请手动配置: %w", err)
	}
	security, ok := value["security"].(map[string]any)
	if !ok {
		security = map[string]any{}
		value["security"] = security
	}
	auth, ok := security["auth"].(map[string]any)
	if !ok {
		auth = map[string]any{}
		security["auth"] = auth
	}
	auth["selectedType"] = "gemini-api-key"
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("编码 Gemini settings.json 失败: %w", err)
	}
	return writeClientFile(path, append(encoded, '\n'))
}

// configureGemini 保持 .env 和 settings.json 的写入顺序与原适配器一致。
func configureGemini(directory, file, endpoint, key string, models []config.ModelEntry, defaultModel string) error {
	if err := configureDotEnvWithModel(file, endpoint, key, "GOOGLE_GEMINI_BASE_URL", "GEMINI_API_KEY", models, defaultModel); err != nil {
		return err
	}
	return configureGeminiSettings(directory + string(os.PathSeparator) + "settings.json")
}
