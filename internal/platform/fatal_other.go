//go:build !windows

/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : 非 Windows 启动错误输出
 */
package platform

import "fmt"

func ShowFatalError(title, message string) {
	fmt.Printf("%s: %s\n", title, message)
}
