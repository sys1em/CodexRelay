/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex 配置文件的地址、模型和本地密钥写入
 * @File          : Codex 客户端适配器
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package clientconfig

import (
	"errors"
	"fmt"
	"os"
)

func configureCodex(configPath, authPath, endpoint, key, defaultModel string) error {
	if err := backupClientFile(configPath); err != nil {
		return err
	}
	if err := backupClientFile(authPath); err != nil {
		return err
	}
	data, err := os.ReadFile(configPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("读取 Codex config.toml 失败: %w", err)
	}
	configText := upsertTomlProviderWithModel(string(data), "codexrelay", endpoint, defaultModel)
	if err := writeClientFile(configPath, []byte(configText)); err != nil {
		return err
	}
	auth, err := readJSONObject(authPath)
	if err != nil {
		return fmt.Errorf("解析 Codex auth.json 失败，请手动配置: %w", err)
	}
	auth["OPENAI_API_KEY"] = key
	return writeJSONObject(authPath, auth)
}
