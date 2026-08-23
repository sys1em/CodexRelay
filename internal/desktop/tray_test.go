/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : 系统托盘令牌筛选回归测试
 */
package desktop

import (
	"testing"

	"codexrelay/internal/config"
)

func TestTrayProfileSelectableFiltersUnavailableDogeTokens(t *testing.T) {
	base := config.Profile{Source: config.SourceDoge, Category: config.CategoryCodex, RemoteTokenID: 42}
	tests := []struct {
		name    string
		profile config.Profile
		tokens  []config.DogeToken
		groups  []string
		want    bool
	}{
		{
			name:    "可用令牌保留",
			profile: base,
			tokens:  []config.DogeToken{{ID: 42, Status: 1, Key: "sk-available", Group: "余额低价组"}},
			groups:  []string{"余额低价组"},
			want:    true,
		},
		{
			name:    "禁用令牌过滤",
			profile: base,
			tokens:  []config.DogeToken{{ID: 42, Status: 2, Group: "余额低价组"}},
			groups:  []string{"余额低价组"},
			want:    false,
		},
		{
			name:    "无权限分组过滤",
			profile: base,
			tokens:  []config.DogeToken{{ID: 42, Status: 1, Group: "余额低价组"}},
			groups:  []string{"余额稳定组"},
			want:    false,
		},
		{
			name:    "目录中已消失过滤",
			profile: base,
			tokens:  []config.DogeToken{{ID: 41, Status: 1, Group: "余额低价组"}},
			groups:  []string{"余额低价组"},
			want:    false,
		},
		{
			name:    "自定义 API 保留",
			profile: config.Profile{Source: config.SourceCustom, Category: config.CategoryCodex, RemoteTokenID: 42},
			tokens:  []config.DogeToken{{ID: 42, Status: 2, Group: "余额低价组"}},
			groups:  nil,
			want:    true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := trayProfileSelectable(test.profile, test.tokens, test.groups); got != test.want {
				t.Fatalf("trayProfileSelectable(%+v, %+v, %v) = %v, want %v", test.profile, test.tokens, test.groups, got, test.want)
			}
		})
	}
}

func TestTrayVisibleCategorySetUsesConfiguredCategories(t *testing.T) {
	visible := trayVisibleCategorySet([]string{config.CategoryCodex, config.CategoryGrok})
	if !visible[config.CategoryCodex] || !visible[config.CategoryGrok] {
		t.Fatalf("configured categories were not retained: %#v", visible)
	}
	if visible[config.CategoryClaude] {
		t.Fatalf("hidden category was included: %#v", visible)
	}
	all := trayVisibleCategorySet(nil)
	if len(all) != len(config.Categories) {
		t.Fatalf("empty visibility should use all categories: got %d want %d", len(all), len(config.Categories))
	}
}

func TestTrayDogeTokenEntryIncludesGroupAndRequiresCompleteKey(t *testing.T) {
	token := config.DogeToken{ID: 42, Category: config.CategoryCodex, Key: "sk-full-key", Status: 1, Group: "余额低价组", GroupDisplayName: "GPT低价组", GroupRatio: 0.02}
	if !trayDogeTokenSelectable(token, []string{"GPT低价组"}) {
		t.Fatal("complete available token should be selectable")
	}
	if trayDogeTokenSelectable(token, []string{"GPT稳定组"}) {
		t.Fatal("token without permitted group should be filtered")
	}
	token.Key = "sk-UkRl**********ZYf4"
	if trayDogeTokenSelectable(token, []string{"GPT低价组"}) {
		t.Fatal("masked token should not be added as an activation entry")
	}
	if got := trayDogeTokenLabel(token.Name, token); got != "未命名令牌（GPT低价组 · 0.02）" {
		t.Fatalf("tray token label = %q", got)
	}
}

func TestDogeTokenDisplayGroupUsesOnlyDisplayName(t *testing.T) {
	token := config.DogeToken{Group: "余额低价组", GroupDisplayName: "GPT低价组"}
	if got := dogeTokenDisplayGroup(token); got != "GPT低价组" {
		t.Fatalf("display group = %q, want display name", got)
	}
	token.GroupDisplayName = ""
	if got := dogeTokenDisplayGroup(token); got != "" {
		t.Fatalf("missing display name should not fall back to raw group: %q", got)
	}
}
