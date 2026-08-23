/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Claude 配置文件的环境变量和模型写入
 * @File          : Claude 客户端适配器
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package clientconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"codexrelay/internal/config"
)

// configureClaude 写入已确认的 Claude 环境变量；同一模型目录的默认项用于四个角色默认值。
func configureClaude(path, endpoint, key string, models []config.ModelEntry, defaultModel string) error {
	if err := backupClientFile(path); err != nil {
		return err
	}
	value := map[string]any{}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &value); err != nil {
			return fmt.Errorf("解析 %s 失败，请手动配置: %w", filepath.Base(path), err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("读取 %s 失败: %w", filepath.Base(path), err)
	}
	env, ok := value["env"].(map[string]any)
	if !ok {
		env = map[string]any{}
		value["env"] = env
	}
	env["ANTHROPIC_BASE_URL"] = endpoint
	env["ANTHROPIC_AUTH_TOKEN"] = key
	if model := selectedModelID(models, defaultModel); model != "" {
		for _, name := range []string{"ANTHROPIC_MODEL", "ANTHROPIC_DEFAULT_HAIKU_MODEL", "ANTHROPIC_DEFAULT_SONNET_MODEL", "ANTHROPIC_DEFAULT_OPUS_MODEL"} {
			env[name] = model
		}
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("编码 %s 失败: %w", filepath.Base(path), err)
	}
	return writeClientFile(path, append(data, '\n'))
}
