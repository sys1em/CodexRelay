/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : 前端静态资源与品牌图标统一嵌入
 */
package codexrelay

import (
	"embed"
	"io/fs"
)

//go:embed frontend logo.png
var resources embed.FS

type frontendFS struct {
	fs.FS
}

func (f frontendFS) Open(name string) (fs.File, error) {
	if name == "logo.png" {
		return resources.Open("logo.png")
	}
	return f.FS.Open(name)
}

func FrontendAssets() (fs.FS, error) {
	assets, err := fs.Sub(resources, "frontend")
	if err != nil {
		return nil, err
	}
	return frontendFS{FS: assets}, nil
}

func AppIcon() []byte {
	icon, _ := resources.ReadFile("logo.png")
	return icon
}
