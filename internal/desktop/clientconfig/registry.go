/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : 外部 AI 客户端路径探测、接管状态检查与受控写入
 * @File          : 外部客户端配置适配器兼容实现
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package clientconfig

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"codexrelay/internal/config"
)

const (
	clientStatusConfigured    = "configured"
	clientStatusNotConfigured = "not_configured"
	clientStatusNotDetected   = "not_detected"
	clientStatusUnsupported   = "unsupported"
	clientStatusError         = "error"
)

// PublicClientConfig 是高级设置和启用前检查使用的脱敏状态，不返回外部配置正文。
type PublicClientConfig struct {
	Category              string `json:"category"`
	Label                 string `json:"label"`
	ConfigDir             string `json:"configDir"`
	ConfigFile            string `json:"configFile"`
	SkipConfigReplacement bool   `json:"skipConfigReplacement"`
	Status                string `json:"status"`
	Detected              bool   `json:"detected"`
	Configured            bool   `json:"configured"`
	StatusText            string `json:"statusText"`
	LastChecked           string `json:"lastChecked"`
	Error                 string `json:"error,omitempty"`
}

type clientDefinition struct {
	Category string
	Label    string
	File     string
	Kind     string
	Default  func() string
}

func clientDefinitions() []clientDefinition {
	home, _ := os.UserHomeDir()
	localAppData := strings.TrimSpace(os.Getenv("LOCALAPPDATA"))
	if localAppData == "" && home != "" {
		localAppData = filepath.Join(home, "AppData", "Local")
	}
	hermesHome := strings.TrimSpace(os.Getenv("HERMES_HOME"))
	if hermesHome == "" {
		hermesHome = filepath.Join(localAppData, "hermes")
	}
	return []clientDefinition{
		{Category: config.CategoryCodex, Label: "Codex", File: "config.toml", Kind: "codex", Default: func() string { return filepath.Join(home, ".codex") }},
		{Category: config.CategoryClaude, Label: "Claude", File: "settings.json", Kind: "claude", Default: func() string { return filepath.Join(home, ".claude") }},
		{Category: config.CategoryGemini, Label: "Gemini", File: ".env", Kind: "gemini", Default: func() string { return filepath.Join(home, ".gemini") }},
		{Category: config.CategoryGrok, Label: "Grok", File: "config.toml", Kind: "grok", Default: func() string { return filepath.Join(home, ".grok") }},
		{Category: config.CategoryOpenCode, Label: "OpenCode", File: "opencode.json", Kind: "opencode", Default: func() string { return filepath.Join(home, ".config", "opencode") }},
		{Category: config.CategoryOpenClaw, Label: "OpenClaw", File: "openclaw.json", Kind: "openclaw", Default: func() string { return filepath.Join(home, ".openclaw") }},
		{Category: config.CategoryHermes, Label: "Hermes", File: "config.yaml", Kind: "hermes", Default: func() string { return hermesHome }},
		{Category: config.CategoryImage, Label: "生图", Kind: "unsupported", Default: func() string { return "" }},
		{Category: config.CategoryOther, Label: "其他", Kind: "unsupported", Default: func() string { return "" }},
	}
}

func clientDefinitionFor(category string) (clientDefinition, bool) {
	for _, definition := range clientDefinitions() {
		if definition.Category == category {
			return definition, true
		}
	}
	return clientDefinition{}, false
}

func clientConfigPath(definition clientDefinition, entry config.ClientConfig) (string, string) {
	directory := strings.TrimSpace(entry.ConfigDir)
	if directory == "" {
		directory = definition.Default()
	}
	filename := strings.TrimSpace(entry.ConfigFile)
	if filename == "" {
		filename = definition.File
	}
	if directory == "" || filename == "" {
		return directory, ""
	}
	return directory, filepath.Join(directory, filename)
}

// discoverClientConfigPaths 只检查各软件已知默认目录，不遍历磁盘；已有自定义路径永远优先保留。
func discoverClientConfigPaths(existing map[string]config.ClientConfig) (map[string]config.ClientConfig, bool) {
	result := make(map[string]config.ClientConfig, len(existing)+len(clientDefinitions()))
	for category, value := range existing {
		result[category] = value
	}
	changed := false
	for _, definition := range clientDefinitions() {
		entry := result[definition.Category]
		if strings.TrimSpace(entry.ConfigDir) != "" || definition.Kind == "unsupported" {
			if entry.ConfigFile == "" && definition.File != "" {
				entry.ConfigFile = definition.File
				result[definition.Category] = entry
				changed = true
			}
			continue
		}
		directory := strings.TrimSpace(definition.Default())
		if directory == "" {
			continue
		}
		_, file := clientConfigPath(definition, config.ClientConfig{ConfigDir: directory, ConfigFile: definition.File})
		if pathExists(directory) || pathExists(file) {
			entry.ConfigDir = directory
			entry.ConfigFile = definition.File
			result[definition.Category] = entry
			changed = true
		}
	}
	return result, changed
}

func pathExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func publicClientConfigs(cfg config.AppConfig) []PublicClientConfig {
	result := make([]PublicClientConfig, 0, len(clientDefinitions()))
	for _, definition := range clientDefinitions() {
		result = append(result, inspectClientConfig(cfg, definition))
	}
	return result
}

func inspectClientConfig(cfg config.AppConfig, definition clientDefinition) PublicClientConfig {
	entry := cfg.ClientConfigs[definition.Category]
	directory, file := clientConfigPath(definition, entry)
	status := PublicClientConfig{Category: definition.Category, Label: definition.Label, ConfigDir: directory, ConfigFile: file, SkipConfigReplacement: entry.SkipConfigReplacement, Status: clientStatusNotDetected, StatusText: "未检测到配置"}
	if definition.Kind == "unsupported" {
		status.Status = clientStatusUnsupported
		status.StatusText = "暂不支持自动配置"
		return status
	}
	if !pathExists(directory) && !pathExists(file) {
		return status
	}
	status.Detected = true
	contents, err := readClientFiles(definition, directory, file)
	if err != nil {
		status.Status = clientStatusError
		status.StatusText = "配置读取失败"
		status.Error = err.Error()
		return status
	}
	endpoint := clientProxyURL(cfg.ProxyPort, definition.Category)
	if clientContentsContain(contents, endpoint, cfg.LocalAccessToken) {
		status.Status = clientStatusConfigured
		status.StatusText = "已由 CodexRelay 配置"
		status.Configured = true
	} else {
		status.Status = clientStatusNotConfigured
		status.StatusText = "未使用 CodexRelay 配置"
	}
	status.LastChecked = time.Now().Format(time.RFC3339)
	return status
}

func clientProxyURL(port int, category string) string {
	if port <= 0 {
		port = config.DefaultProxyPort
	}
	return fmt.Sprintf("http://127.0.0.1:%d/%s", port, category)
}

func readClientFiles(definition clientDefinition, directory, file string) ([][]byte, error) {
	paths := []string{file}
	if definition.Kind == "codex" {
		paths = []string{file, filepath.Join(directory, "auth.json")}
	} else if definition.Kind == "gemini" {
		paths = []string{filepath.Join(directory, ".env"), filepath.Join(directory, "settings.json")}
	}
	contents := make([][]byte, 0, len(paths))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("读取 %s: %w", filepath.Base(path), err)
		}
		contents = append(contents, data)
	}
	return contents, nil
}

func clientContentsContain(contents [][]byte, endpoint, key string) bool {
	if endpoint == "" || key == "" {
		return false
	}
	foundEndpoint := false
	foundKey := false
	for _, content := range contents {
		foundEndpoint = foundEndpoint || bytes.Contains(content, []byte(endpoint))
		foundKey = foundKey || bytes.Contains(content, []byte(key))
	}
	return foundEndpoint && foundKey
}

// configureClient 只使用已保存代理 API 的模型目录和当前监听信息；不会在配置外部客户端时联网获取模型。
func configureClient(cfg config.AppConfig, category, profileID string) error {
	definition, ok := clientDefinitionFor(category)
	if !ok || definition.Kind == "unsupported" {
		return errors.New("该 API 类别暂不支持自动配置，请手动配置")
	}
	entry := cfg.ClientConfigs[category]
	directory, file := clientConfigPath(definition, entry)
	if strings.TrimSpace(directory) == "" || strings.TrimSpace(file) == "" {
		return errors.New("请先在高级设置中设置配置目录")
	}
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("创建配置目录: %w", err)
	}
	endpoint := clientProxyURL(cfg.ProxyPort, category)
	key := strings.TrimSpace(cfg.LocalAccessToken)
	profile := activeProfileForClient(cfg, category, profileID)
	var models []config.ModelEntry
	defaultModel := ""
	if profile != nil {
		models = profile.Models
		defaultModel = profile.DefaultModel
	}
	switch definition.Kind {
	case "claude":
		return configureClaude(file, endpoint, key, models, defaultModel)
	case "gemini":
		if err := configureDotEnvWithModel(filepath.Join(directory, ".env"), endpoint, key, "GOOGLE_GEMINI_BASE_URL", "GEMINI_API_KEY", models, defaultModel); err != nil {
			return err
		}
		return configureGeminiSettings(filepath.Join(directory, "settings.json"))
	case "opencode":
		return configureOpenCode(file, endpoint, key, models, defaultModel)
	case "openclaw":
		return configureOpenClaw(file, endpoint, key, models, defaultModel)
	case "codex":
		if err := configureCodex(file, filepath.Join(directory, "auth.json"), endpoint, key, defaultModel); err != nil {
			return err
		}
		return nil
	case "grok":
		return configureGrok(file, endpoint, key, models, defaultModel)
	case "hermes":
		return configureHermes(file, endpoint, key, models, defaultModel)
	default:
		return errors.New("该 API 类别暂不支持自动配置，请手动配置")
	}
}

// DiscoverConfigPaths 只探测已知默认目录，并保留用户已经保存的自定义目录。
func DiscoverConfigPaths(existing map[string]config.ClientConfig) (map[string]config.ClientConfig, bool) {
	return discoverClientConfigPaths(existing)
}

// PublicConfigs 返回所有客户端的脱敏状态；它只读取本地配置文件。
func PublicConfigs(cfg config.AppConfig) []PublicClientConfig {
	return publicClientConfigs(cfg)
}

// Inspect 返回指定分类的本地配置状态，不会访问上游服务。
func Inspect(cfg config.AppConfig, category string) (PublicClientConfig, error) {
	definition, ok := clientDefinitionFor(category)
	if !ok {
		return PublicClientConfig{}, errors.New("未知 API 类别")
	}
	return inspectClientConfig(cfg, definition), nil
}

// Configure 备份并写入指定外部客户端配置；配置内容来自当前本地快照。
func Configure(cfg config.AppConfig, category, profileID string) error {
	return configureClient(cfg, category, profileID)
}

// ConfigFileFor 返回分类的默认配置文件名，用于保存高级设置中的自定义目录。
func ConfigFileFor(category string) (string, bool) {
	definition, ok := clientDefinitionFor(category)
	if !ok {
		return "", false
	}
	return definition.File, true
}

// Supports 表示分类是否有已实现的自动配置适配器。
func Supports(category string) bool {
	definition, ok := clientDefinitionFor(category)
	return ok && definition.Kind != "unsupported"
}
