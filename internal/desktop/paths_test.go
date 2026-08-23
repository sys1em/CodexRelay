/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : CodexRelay 路径选择结果回归测试
 * @File          : 路径选择测试
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package desktop

import "testing"

func TestNormalizeSelectedDirectoryKeepsCancelEmpty(t *testing.T) {
	if got := normalizeSelectedDirectory(""); got != "" {
		t.Fatalf("empty selection = %q", got)
	}
	if got := normalizeSelectedDirectory("  "); got != "" {
		t.Fatalf("whitespace selection = %q", got)
	}
}
