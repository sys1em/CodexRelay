//go:build !windows

/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : 非 Windows 自启动占位实现
 */
package platform

func SetLaunchAtStartup(_ bool) error { return nil }

func StartupCommand(executable string) string { return executable + " --autostart" }
