/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : 系统托盘菜单与窗口恢复
 */
package desktop

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"codexrelay/internal/config"
	"codexrelay/internal/relay"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func setupTray(
	wailsApp *application.App,
	window *application.WebviewWindow,
	runtime *relay.Runtime,
	service *DesktopService,
	icon []byte,
	restoreWindow func(),
) {
	tray := wailsApp.SystemTray.New()
	tray.SetIcon(icon)
	tray.SetDarkModeIcon(icon)
	tray.SetTooltip("CodexRelay - 代理 API 运行中")
	tray.OnClick(restoreWindow)
	tray.OnDoubleClick(restoreWindow)

	refreshTray := func() {
		state := runtime.State()
		if state == nil {
			return
		}
		menu := wailsApp.NewMenu()
		activeName := "未选择"
		if active := state.Active[config.CategoryCodex]; active != nil {
			activeName = active.Profile.Name
		}
		menu.Add("打开").OnClick(func(_ *application.Context) { restoreWindow() })
		// 类别直接作为托盘一级菜单，减少用户展开层级；隐藏类别和空类别不创建菜单项。
		visibleCategories := trayVisibleCategorySet(state.Config.Preferences.VisibleCategories)
		profilesByTokenID := make(map[int64]config.Profile)
		for _, profile := range state.Config.Profiles {
			if profile.Source == config.SourceDoge && profile.RemoteTokenID > 0 {
				profilesByTokenID[profile.RemoteTokenID] = profile
			}
		}
		tokensByID := make(map[int64]config.DogeToken, len(state.Config.Doge.Tokens))
		for _, token := range state.Config.Doge.Tokens {
			tokensByID[token.ID] = token
		}
		for _, category := range config.Categories {
			if !visibleCategories[category] {
				continue
			}
			entries := make([]trayMenuEntry, 0)
			seenTokenIDs := make(map[int64]struct{})
			for _, profile := range state.Config.Profiles {
				if profile.Category != category || !trayProfileSelectable(profile, state.Config.Doge.Tokens, state.Config.Doge.Groups) {
					continue
				}
				entry := trayMenuEntry{category: category, profileID: profile.ID, name: profile.Name, current: state.Config.ActiveProfiles[category] == profile.ID}
				if profile.Source == config.SourceDoge && profile.RemoteTokenID > 0 {
					if token, ok := tokensByID[profile.RemoteTokenID]; ok {
						entry.name = trayDogeTokenLabel(profile.Name, token)
					}
					seenTokenIDs[profile.RemoteTokenID] = struct{}{}
				}
				entries = append(entries, entry)
			}
			availableTokens := availableDogeTokensForCategory(state.Config.Doge.Tokens, profilesByTokenID, state.Config.Doge.Groups, category)
			for _, token := range availableTokens {
				if _, imported := profilesByTokenID[token.ID]; imported {
					continue
				}
				if _, seen := seenTokenIDs[token.ID]; seen {
					continue
				}
				entries = append(entries, trayMenuEntry{category: category, tokenID: token.ID, name: trayDogeTokenLabel(token.Name, token)})
			}
			if len(entries) == 0 {
				continue
			}
			categoryMenu := menu.AddSubmenu(categoryLabel(category))
			for _, entry := range entries {
				entry := entry
				categoryMenu.AddRadio(entry.name, entry.current).OnClick(func(_ *application.Context) {
					var err error
					if entry.profileID != "" {
						err = service.ActivateProfile(entry.profileID)
					} else {
						err = service.EnableDogeToken(entry.tokenID)
					}
					if err != nil {
						wailsApp.Logger.Error("托盘切换代理 API 失败", "error", err)
					}
				})
			}
		}
		menu.AddSeparator()
		menu.Add("退出").OnClick(func(_ *application.Context) { wailsApp.Quit() })
		tray.SetMenu(menu)
		tray.SetTooltip(fmt.Sprintf("CodexRelay - %s", activeName))
		window.EmitEvent("relay-state-changed")
	}
	// 原生托盘菜单更新可能阻塞 Wails 调用；串行异步刷新并合并中间状态，避免切换 API 时卡住前端。
	refreshMu := sync.Mutex{}
	refreshing := false
	queued := false
	scheduleRefresh := func() {
		refreshMu.Lock()
		if refreshing {
			queued = true
			refreshMu.Unlock()
			return
		}
		refreshing = true
		refreshMu.Unlock()
		go func() {
			for {
				refreshTray()
				refreshMu.Lock()
				if !queued {
					refreshing = false
					refreshMu.Unlock()
					return
				}
				queued = false
				refreshMu.Unlock()
			}
		}()
	}
	service.addStateChangedHandler(scheduleRefresh)
	refreshTray()
}

type trayMenuEntry struct {
	category  string
	profileID string
	tokenID   int64
	name      string
	current   bool
}

func trayVisibleCategorySet(categories []string) map[string]bool {
	visible := make(map[string]bool, len(config.Categories))
	if len(categories) == 0 {
		for _, category := range config.Categories {
			visible[category] = true
		}
		return visible
	}
	for _, category := range categories {
		if config.IsCategory(category) {
			visible[category] = true
		}
	}
	return visible
}

func trayDogeTokenSelectable(token config.DogeToken, groups []string) bool {
	return len(availableDogeTokensForCategory([]config.DogeToken{token}, nil, groups, token.Category)) == 1
}

func trayDogeTokenLabel(name string, token config.DogeToken) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "未命名令牌"
	}
	group := dogeTokenDisplayGroup(token)
	if group == "" {
		return name
	}
	if token.GroupRatio > 0 {
		return fmt.Sprintf("%s（%s · %s）", name, group, formatDogeRatio(token.GroupRatio))
	}
	return fmt.Sprintf("%s（%s）", name, group)
}

func formatDogeRatio(ratio float64) string {
	return strconv.FormatFloat(ratio, 'f', -1, 64)
}

// trayProfileSelectable 判断托盘切换菜单是否应展示某个 Profile。
// 自定义 API 没有二狗子目录状态；二狗子 API 复用主窗口类别可用集合，要求令牌状态、分组权限和本地完整密钥均有效。
// 本地目录缺少令牌时直接隐藏，避免菜单展示一个后端无法启用的目标。
func trayProfileSelectable(profile config.Profile, tokens []config.DogeToken, groups []string) bool {
	if profile.Source != config.SourceDoge || profile.RemoteTokenID <= 0 {
		return true
	}
	for _, token := range tokens {
		if token.ID != profile.RemoteTokenID {
			continue
		}
		profilesByRemoteID := map[int64]config.Profile{profile.RemoteTokenID: profile}
		return len(availableDogeTokensForCategory([]config.DogeToken{token}, profilesByRemoteID, groups, profile.Category)) == 1
	}
	return false
}

func categoryLabel(category string) string {
	labels := map[string]string{
		config.CategoryCodex:    "Codex",
		config.CategoryClaude:   "Claude",
		config.CategoryGemini:   "Gemini",
		config.CategoryGrok:     "Grok",
		config.CategoryOpenCode: "OpenCode",
		config.CategoryOpenClaw: "OpenClaw",
		config.CategoryHermes:   "Hermes",
		config.CategoryImage:    "生图",
		config.CategoryOther:    "其他",
	}
	if label, ok := labels[category]; ok {
		return label
	}
	return category
}
