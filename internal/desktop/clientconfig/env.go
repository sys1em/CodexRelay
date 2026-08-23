/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : .env 配置中的地址、密钥和模型字段更新
 * @File          : 环境变量配置辅助
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package clientconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"codexrelay/internal/config"
)

func configureDotEnv(path, endpoint, key, baseKey, tokenKey string) error {
	return configureDotEnvWithModel(path, endpoint, key, baseKey, tokenKey, nil, "")
}

func configureDotEnvWithModel(path, endpoint, key, baseKey, tokenKey string, models []config.ModelEntry, defaultModel string) error {
	if err := backupClientFile(path); err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("读取 %s 失败: %w", filepath.Base(path), err)
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	lines = upsertEnvLine(lines, baseKey, endpoint)
	lines = upsertEnvLine(lines, tokenKey, key)
	if model := selectedModelID(models, defaultModel); model != "" {
		modelKey := "GEMINI_MODEL"
		if baseKey != "GOOGLE_GEMINI_BASE_URL" {
			modelKey = "MODEL"
		}
		lines = upsertEnvLine(lines, modelKey, model)
	}
	content := strings.TrimRight(strings.Join(lines, "\n"), "\n") + "\n"
	return writeClientFile(path, []byte(content))
}

func upsertEnvLine(lines []string, key, value string) []string {
	prefix := key + "="
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			lines[i] = prefix + value
			return lines
		}
	}
	return append(lines, prefix+value)
}
