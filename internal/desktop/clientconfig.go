/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : 桌面层客户端配置绑定与配置持久化调度
 * @File          : 外部客户端配置桌面入口
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package desktop

import (
	"errors"
	"path/filepath"
	"strings"

	"codexrelay/internal/config"
	"codexrelay/internal/desktop/clientconfig"
)

// PublicClientConfig 是高级设置和启用前检查使用的脱敏状态，不返回外部配置正文。
// 类型别名保持 Wails 绑定中的公开字段和 JSON 形状不变。
type PublicClientConfig = clientconfig.PublicClientConfig

func publicClientConfigs(cfg config.AppConfig) []PublicClientConfig {
	return clientconfig.PublicConfigs(cfg)
}

// scanClientConfigs 在启动阶段记录已安装客户端的已知配置目录；失败只返回错误，不修改外部客户端文件。
func (s *DesktopService) scanClientConfigs() error {
	state := s.runtime.State()
	if state == nil {
		return errors.New("程序尚未初始化")
	}
	paths, changed := clientconfig.DiscoverConfigPaths(state.Config.ClientConfigs)
	if !changed {
		return nil
	}
	return s.updateConfig(func(cfg *config.AppConfig) error {
		cfg.ClientConfigs = paths
		return nil
	})
}

// CheckClientConfig 返回某一分类当前配置文件的实时状态；它只读本地文件，不访问上游。
func (s *DesktopService) CheckClientConfig(category string) (PublicClientConfig, error) {
	state := s.runtime.State()
	if state == nil {
		return PublicClientConfig{}, errors.New("程序尚未初始化")
	}
	return clientconfig.Inspect(state.Config, strings.TrimSpace(category))
}

// SetClientConfigPath 保存外部客户端配置目录，并立即返回新的本地检测状态。
func (s *DesktopService) SetClientConfigPath(category, directory string) error {
	category = strings.TrimSpace(category)
	directory = strings.TrimSpace(directory)
	configFile, ok := clientconfig.ConfigFileFor(category)
	if !ok {
		return errors.New("未知 API 类别")
	}
	if !clientconfig.Supports(category) {
		return errors.New("该 API 类别没有可自动配置的外部客户端")
	}
	if directory == "" {
		return errors.New("配置目录不能为空")
	}
	if !filepath.IsAbs(directory) {
		return errors.New("配置目录必须是绝对路径")
	}
	return s.updateConfig(func(cfg *config.AppConfig) error {
		if cfg.ClientConfigs == nil {
			cfg.ClientConfigs = map[string]config.ClientConfig{}
		}
		entry := cfg.ClientConfigs[category]
		entry.ConfigDir = filepath.Clean(directory)
		entry.ConfigFile = configFile
		cfg.ClientConfigs[category] = entry
		return nil
	})
}

// SetClientConfigSkip 保存某一外部客户端是否跳过主界面切换前的配置检查；它不改写外部文件。
func (s *DesktopService) SetClientConfigSkip(category string, skip bool) error {
	category = strings.TrimSpace(category)
	if !clientconfig.Supports(category) {
		return errors.New("该 API 类别没有可自动配置的外部客户端")
	}
	return s.updateConfig(func(cfg *config.AppConfig) error {
		if cfg.ClientConfigs == nil {
			cfg.ClientConfigs = map[string]config.ClientConfig{}
		}
		entry := cfg.ClientConfigs[category]
		entry.SkipConfigReplacement = skip
		if entry.ConfigFile == "" {
			if filename, ok := clientconfig.ConfigFileFor(category); ok {
				entry.ConfigFile = filename
			}
		}
		cfg.ClientConfigs[category] = entry
		return nil
	})
}

// ConfigureClient 备份并写入当前分类的 CodexRelay 地址和本地访问令牌。
// 它不改变令牌启用映射，调用方完成写入后再执行启用，避免配置失败却显示为已切换。
func (s *DesktopService) ConfigureClient(category, profileID string) error {
	state := s.runtime.State()
	if state == nil {
		return errors.New("程序尚未初始化")
	}
	if err := clientconfig.Configure(state.Config, strings.TrimSpace(category), strings.TrimSpace(profileID)); err != nil {
		return err
	}
	s.notifyStateChanged()
	return nil
}
