//go:build windows

/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : Windows 开机自启动注册
 */
package platform

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const startupRegistryPath = `Software\Microsoft\Windows\CurrentVersion\Run`

func SetLaunchAtStartup(enabled bool) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, startupRegistryPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	if !enabled {
		err = key.DeleteValue("CodexRelay")
		if err == registry.ErrNotExist {
			return nil
		}
		return err
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	executable = strings.TrimSpace(executable)
	if executable == "" {
		return fmt.Errorf("无法确定程序路径")
	}
	return key.SetStringValue("CodexRelay", StartupCommand(executable))
}

func StartupCommand(executable string) string {
	return `"` + executable + `" --autostart`
}
