/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : 唯一桌面程序入口与启动参数
 */
package main

import (
	"flag"
	"os"

	"codexrelay/internal/desktop"
	"codexrelay/internal/platform"
)

func main() {
	var options desktop.LaunchOptions
	flag.IntVar(&options.ProxyPort, "proxy-port", 0, "覆盖透明代理端口")
	flag.BoolVar(&options.Autostart, "autostart", false, "标记为 Windows 开机自启动")
	flag.Parse()
	if err := desktop.Run(options); err != nil {
		platform.ShowFatalError("CodexRelay 启动失败", err.Error())
		os.Exit(1)
	}
}
