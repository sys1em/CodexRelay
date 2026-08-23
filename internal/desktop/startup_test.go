/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : 桌面启动来源与隐藏语义测试
 */
package desktop

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codexrelay/internal/config"
	"codexrelay/internal/platform"
)

func TestWindowOnlyStartsHiddenFromAutostart(t *testing.T) {
	preferences := config.Preferences{StartHidden: true}
	if shouldStartHidden(false, preferences) {
		t.Fatal("manual launch must remain visible")
	}
	if !shouldStartHidden(true, preferences) {
		t.Fatal("autostart launch should honor hidden preference")
	}
	if shouldStartHidden(true, config.Preferences{}) {
		t.Fatal("autostart should remain visible when preference is disabled")
	}
}

func TestStartupCommandMarksAutostartLaunch(t *testing.T) {
	command := platform.StartupCommand(`C:\Program Files\CodexRelay\CodexRelay.exe`)
	if !strings.Contains(command, "--autostart") {
		t.Fatalf("startup command = %q", command)
	}
}

func TestOnboardingRequiredWhenPortableDataFileMissing(t *testing.T) {
	directory := t.TempDir()
	configPath := filepath.Join(directory, "config.json")
	usagePath := filepath.Join(directory, "usage.json")

	if required, err := onboardingRequired(configPath, usagePath); err != nil || !required {
		t.Fatalf("both missing files should require onboarding, required=%v err=%v", required, err)
	}
	if err := os.WriteFile(configPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if required, err := onboardingRequired(configPath, usagePath); err != nil || !required {
		t.Fatalf("missing usage file should require onboarding, required=%v err=%v", required, err)
	}
	if err := os.WriteFile(usagePath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if required, err := onboardingRequired(configPath, usagePath); err != nil || required {
		t.Fatalf("both existing files should skip onboarding, required=%v err=%v", required, err)
	}
	if err := os.Remove(configPath); err != nil {
		t.Fatal(err)
	}
	if required, err := onboardingRequired(configPath, usagePath); err != nil || !required {
		t.Fatalf("missing config file should require onboarding, required=%v err=%v", required, err)
	}
}
