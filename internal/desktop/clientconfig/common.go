/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : 外部客户端配置适配器共用的备份、写入和模型选择逻辑
 * @File          : 客户端配置公共辅助
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package clientconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"codexrelay/internal/config"
	"codexrelay/internal/storage"
)

// activeProfileForClient 从本地配置快照选择指定分类的模型目录，不触发网络请求。
func activeProfileForClient(cfg config.AppConfig, category, profileID string) *config.Profile {
	if profileID == "" {
		profileID = cfg.ActiveProfiles[category]
	}
	index := config.FindProfileIndex(cfg.Profiles, profileID)
	if index < 0 || cfg.Profiles[index].Category != category {
		return nil
	}
	profile := config.CloneProfile(cfg.Profiles[index])
	return &profile
}

// backupClientFile 为外部文件创建带时间戳的 .CodexRelay 备份，不存在的文件不生成空备份。
func backupClientFile(path string) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("读取备份源文件: %w", err)
	}
	stamp := time.Now().Format("20060102-150405")
	backup := fmt.Sprintf("%s.%s.CodexRelay", path, stamp)
	for index := 2; pathExists(backup); index++ {
		backup = fmt.Sprintf("%s.%s-%d.CodexRelay", path, stamp, index)
	}
	if err := storage.WriteBytesAtomic(backup, ".codexrelay-backup-*.tmp", data, 0o600); err != nil {
		return fmt.Errorf("创建 %s 备份: %w", filepath.Base(path), err)
	}
	return nil
}

// writeClientFile 通过共享的原子 JSON 存储写入外部配置，失败时保留原文件。
func writeClientFile(path string, data []byte) error {
	return storage.WriteBytesAtomic(path, ".codexrelay-config-*.tmp", data, 0o600)
}

func containsModel(models []config.ModelEntry, id string) bool {
	for _, model := range models {
		if model.ID == id {
			return true
		}
	}
	return false
}

func selectedModelID(models []config.ModelEntry, defaultModel string) string {
	if defaultModel != "" && containsModel(models, defaultModel) {
		return defaultModel
	}
	if len(models) > 0 {
		return models[0].ID
	}
	return ""
}
