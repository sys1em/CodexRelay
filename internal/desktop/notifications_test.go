/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : 右下角提醒窗口启动状态回归测试
 * @File          : 提醒管理器就绪条件测试
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package desktop

import "testing"

func TestNotificationManagerStartsWithoutMainWindowReady(t *testing.T) {
	manager := &notificationWindowManager{}
	manager.markStarted()

	manager.mu.Lock()
	started := manager.started
	manager.mu.Unlock()
	if !started {
		t.Fatal("提醒窗口就绪后管理器仍未进入可刷新状态")
	}
}
