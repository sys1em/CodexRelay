/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : 右下角公告、额度和令牌切换提醒窗口管理
 * @File          : 独立 Wails 提醒窗口与主窗口状态联动
 */
package desktop

import (
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

const (
	notificationWindowWidth  = 430
	notificationWindowHeight = 300
	notificationWindowMargin = 18
	notificationWindowGap    = 12
)

var notificationKinds = []string{
	NotificationKindBalance,
	NotificationKindSubscription,
	NotificationKindAnnouncement,
	NotificationKindTokenSwitch,
}

func notificationWindowSize() (int, int) {
	return notificationWindowWidth, notificationWindowHeight
}

// notificationWindowManager 按提醒类别维护独立窗口；窗口只在主窗口隐藏或最小化时显示。
type notificationWindowManager struct {
	mu         sync.Mutex
	mainWindow *application.WebviewWindow
	service    *DesktopService
	windows    map[string]*application.WebviewWindow
	started    bool
}

func newNotificationWindowManager(wailsApp *application.App, mainWindow *application.WebviewWindow, service *DesktopService) *notificationWindowManager {
	manager := &notificationWindowManager{
		mainWindow: mainWindow,
		service:    service,
		windows:    make(map[string]*application.WebviewWindow, len(notificationKinds)),
	}
	for _, kind := range notificationKinds {
		width, height := notificationWindowSize()
		window := wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
			Name:             "notification-" + kind,
			Title:            "CodexRelay 提醒",
			Width:            width,
			Height:           height,
			MinWidth:         width,
			MinHeight:        height,
			MaxWidth:         width,
			MaxHeight:        height,
			URL:              "/notification.html?kind=" + kind,
			Frameless:        true,
			DisableResize:    true,
			AlwaysOnTop:      true,
			Hidden:           true,
			BackgroundColour: application.NewRGBA(255, 255, 255, 0),
		})
		manager.windows[kind] = window
	}
	mainWindow.RegisterHook(events.Common.WindowRuntimeReady, func(_ *application.WindowEvent) {
		manager.mu.Lock()
		manager.started = true
		manager.mu.Unlock()
		manager.refresh()
	})
	for _, eventType := range []events.WindowEventType{events.Common.WindowShow, events.Common.WindowRestore, events.Common.WindowUnMinimise} {
		mainWindow.RegisterHook(eventType, func(_ *application.WindowEvent) { manager.hideAll() })
	}
	for _, eventType := range []events.WindowEventType{events.Common.WindowMinimise, events.Common.WindowHide} {
		mainWindow.RegisterHook(eventType, func(_ *application.WindowEvent) { manager.refresh() })
	}
	return manager
}

func (m *notificationWindowManager) refresh() {
	m.mu.Lock()
	started := m.started
	m.mu.Unlock()
	if !started || m.mainWindow == nil {
		return
	}
	if m.mainWindow.IsVisible() && !m.mainWindow.IsMinimised() {
		m.hideAll()
		return
	}
	state := m.service.GetState()
	alertsByKind := make(map[string][]PublicDogeAlert, len(notificationKinds))
	for _, alert := range state.Doge.Notifications.Alerts {
		alertsByKind[alert.Kind] = append(alertsByKind[alert.Kind], alert)
	}
	if state.Doge.TokenSwitch != nil {
		alertsByKind[NotificationKindTokenSwitch] = []PublicDogeAlert{{Kind: NotificationKindTokenSwitch, Key: state.Doge.TokenSwitch.Key}}
	}
	screen, _ := m.mainWindow.GetScreen()
	stackOffset := 0
	for _, kind := range notificationKinds {
		window := m.windows[kind]
		width, height := notificationWindowSize()
		alerts := alertsByKind[kind]
		if len(alerts) == 0 {
			window.Hide()
			continue
		}
		if screen != nil {
			window.SetScreen(screen)
			workArea := screen.WorkArea
			x := workArea.X + workArea.Width - width - notificationWindowMargin
			y := workArea.Y + workArea.Height - height - notificationWindowMargin - stackOffset
			window.SetPosition(x, y)
		}
		window.EmitEvent("notification-state-changed")
		window.Show()
		stackOffset += height + notificationWindowGap
	}
}

func (m *notificationWindowManager) hideAll() {
	for _, window := range m.windows {
		window.Hide()
	}
}
