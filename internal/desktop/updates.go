/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Windows 版本检测和更新公开模型
 * @File          : 桌面更新公共模型
 */
package desktop

type UpdateInfo struct {
	Supported      bool   `json:"supported"`
	Available      bool   `json:"available"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion,omitempty"`
	Name           string `json:"name,omitempty"`
	PublishedAt    string `json:"publishedAt,omitempty"`
	Size           int64  `json:"size,omitempty"`
}
