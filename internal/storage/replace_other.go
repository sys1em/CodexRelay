//go:build !windows

/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : 非 Windows 原子文件替换
 */
package storage

import "os"

func replaceFile(source, destination string) error {
	return os.Rename(source, destination)
}
