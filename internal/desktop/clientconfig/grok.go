/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Grok TOML 模型段和访问凭据写入
 * @File          : Grok 客户端适配器
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package clientconfig

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"codexrelay/internal/config"
)

func configureGrok(path, endpoint, key string, models []config.ModelEntry, defaultModel string) error {
	if len(models) == 0 {
		return errors.New("Grok 配置需要至少一个模型，请先在编辑页获取模型列表")
	}
	if err := backupClientFile(path); err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("读取 Grok config.toml 失败: %w", err)
	}
	raw := string(data)
	if strings.TrimSpace(raw) == "" {
		raw = ""
	}
	if defaultModel == "" || !containsModel(models, defaultModel) {
		defaultModel = models[0].ID
	}
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	lines = upsertTomlSectionValue(lines, "models", "default", strconv.Quote(defaultModel))
	for _, model := range models {
		name := strings.TrimSpace(model.Name)
		if name == "" {
			name = model.ID
		}
		contextWindow := model.ContextWindow
		if contextWindow <= 0 {
			// Grok 要求正的 context_window；该默认值来自已确认的客户端适配器。
			contextWindow = 500000
		}
		header := "model." + strconv.Quote(model.ID)
		section := []string{
			"model = " + strconv.Quote(model.ID),
			"name = " + strconv.Quote(name),
			"base_url = " + strconv.Quote(endpoint),
			"api_key = " + strconv.Quote(key),
			"api_backend = \"responses\"",
			"context_window = " + strconv.FormatInt(contextWindow, 10),
		}
		lines = upsertTomlSection(lines, header, section)
	}
	return writeClientFile(path, []byte(strings.TrimRight(strings.Join(lines, "\n"), "\n")+"\n"))
}
