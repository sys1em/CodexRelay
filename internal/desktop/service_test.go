/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : 桌面绑定服务回归测试
 */
package desktop

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"codexrelay/internal/config"
	"codexrelay/internal/relay"
	"codexrelay/internal/usage"
)

func TestReorderProfilesPersistsValidatedOrder(t *testing.T) {
	directory := t.TempDir()
	configPath := filepath.Join(directory, "config.json")
	store := config.NewStore(configPath)
	cfg := config.Default(18765)
	cfg.Profiles = []config.Profile{
		{ID: "one", Source: config.SourceCustom, Category: config.CategoryCodex, Name: "One", BaseURL: "https://one.example/v1", APIKey: "sk-one"},
		{ID: "two", Source: config.SourceCustom, Category: config.CategoryCodex, Name: "Two", BaseURL: "https://two.example/v1", APIKey: "sk-two"},
		{ID: "three", Source: config.SourceCustom, Category: config.CategoryCodex, Name: "Three", BaseURL: "https://three.example/v1", APIKey: "sk-three"},
	}
	if err := store.Save(cfg); err != nil {
		t.Fatal(err)
	}
	runtime := newTestRuntime(t, directory, store, cfg)
	service := NewDesktopService(runtime)
	if err := service.ReorderProfiles([]string{"three", "one", "two"}); err != nil {
		t.Fatal(err)
	}
	state := runtime.State()
	if state.Config.Profiles[0].ID != "three" || state.Config.Profiles[2].ID != "two" {
		t.Fatalf("runtime order = %+v", state.Config.Profiles)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var persisted config.AppConfig
	if err := json.Unmarshal(data, &persisted); err != nil {
		t.Fatal(err)
	}
	if persisted.Profiles[0].ID != "three" || persisted.Profiles[1].ID != "one" || persisted.Profiles[2].ID != "two" {
		t.Fatalf("persisted order = %+v", persisted.Profiles)
	}
	if err := service.ReorderProfiles([]string{"three", "three", "two"}); err == nil {
		t.Fatal("duplicate profile order should fail")
	}
}

func TestSaveProfileReturnsPlaintextKeyAndNote(t *testing.T) {
	directory := t.TempDir()
	store := config.NewStore(filepath.Join(directory, "config.json"))
	cfg := config.Default(18765)
	if err := store.Save(cfg); err != nil {
		t.Fatal(err)
	}
	runtime := newTestRuntime(t, directory, store, cfg)
	service := NewDesktopService(runtime)
	input := ProfileInput{Source: config.SourceCustom, Category: config.CategoryCodex, Name: "One", BaseURL: "https://example.test/v1", APIKey: "sk-example", Note: "稳定线路"}
	if err := service.SaveProfile(input); err != nil {
		t.Fatal(err)
	}
	state := service.GetState()
	if len(state.Profiles) != 1 || state.Profiles[0].APIKey != "sk-example" || state.Profiles[0].Note != "稳定线路" {
		t.Fatalf("profiles = %+v", state.Profiles)
	}
}

func TestSaveProfileKeepsDogeKeyImmutable(t *testing.T) {
	directory := t.TempDir()
	store := config.NewStore(filepath.Join(directory, "config.json"))
	cfg := config.Default(18765)
	cfg.Profiles = []config.Profile{{
		ID: "doge-profile", Source: config.SourceDoge, Category: config.CategoryCodex,
		Name: "原名称", BaseURL: "https://example.test/v1", APIKey: "sk-original", RemoteTokenID: 42,
	}}
	if err := store.Save(cfg); err != nil {
		t.Fatal(err)
	}
	runtime := newTestRuntime(t, directory, store, cfg)
	service := NewDesktopService(runtime)

	if err := service.SaveProfile(ProfileInput{
		ID: "doge-profile", Source: config.SourceDoge, Category: config.CategoryCodex,
		Name: "修改名称", BaseURL: "https://example.test/v1", APIKey: "sk-changed",
	}); err == nil {
		t.Fatal("changing a doge API key should fail")
	}
	if err := service.SaveProfile(ProfileInput{
		ID: "doge-profile", Source: config.SourceDoge, Category: config.CategoryCodex,
		Name: "修改名称", BaseURL: "https://example.test/v1", APIKey: "sk-original", Note: "可编辑备注",
	}); err != nil {
		t.Fatalf("saving other doge fields = %v", err)
	}

	profile := runtime.State().Config.Profiles[0]
	if profile.Name != "修改名称" || profile.Note != "可编辑备注" || profile.APIKey != "sk-original" {
		t.Fatalf("profile = %+v", profile)
	}
}

func TestUnbindDogeClearsAccountSnapshot(t *testing.T) {
	directory := t.TempDir()
	store := config.NewStore(filepath.Join(directory, "config.json"))
	cfg := config.Default(18765)
	cfg.Doge.AccessToken = "fake-doge-access-token"
	cfg.Doge.Account = config.DogeAccount{ID: 7, Quota: 1000000}
	cfg.Doge.Subscriptions = []config.DogeSubscription{{ID: 9, Status: "active", AmountTotal: 2000000}}
	cfg.Doge.Topup = config.DogeTopupInfo{EnableRedemption: true, TopupLink: "https://example.test/buy"}
	cfg.Doge.User = map[string]any{"username": "fake-user"}
	if err := store.Save(cfg); err != nil {
		t.Fatal(err)
	}
	runtime := newTestRuntime(t, directory, store, cfg)
	service := NewDesktopService(runtime)
	if err := service.UnbindDoge(); err != nil {
		t.Fatal(err)
	}
	doge := runtime.State().Config.Doge
	if doge.AccessToken != "" || doge.Account.Quota != 0 || len(doge.Subscriptions) != 0 || doge.Topup.EnableRedemption || doge.Topup.TopupLink != "" || doge.User != nil {
		t.Fatalf("unbound doge snapshot = %+v", doge)
	}
}

func TestCompleteOnboardingClearsRuntimeFlag(t *testing.T) {
	directory := t.TempDir()
	store := config.NewStore(filepath.Join(directory, "config.json"))
	cfg := config.Default(18765)
	if err := store.Save(cfg); err != nil {
		t.Fatal(err)
	}
	runtime := newTestRuntime(t, directory, store, cfg)
	service := NewDesktopService(runtime)
	service.setNeedsOnboarding(true)
	if !service.GetState().NeedsOnboarding {
		t.Fatal("onboarding flag should be visible before completion")
	}
	if err := service.CompleteOnboarding(); err != nil {
		t.Fatal(err)
	}
	if service.GetState().NeedsOnboarding {
		t.Fatal("onboarding flag should be cleared after completion")
	}
}

func TestCompleteOnboardingActivatesDeferredPortableStores(t *testing.T) {
	directory := t.TempDir()
	configPath := filepath.Join(directory, "config.json")
	usagePath := filepath.Join(directory, "usage.json")
	store := config.NewDeferredStore(configPath)
	cfg, err := store.LoadOrCreate(18765)
	if err != nil {
		t.Fatal(err)
	}
	usageStore, err := usage.NewDeferredStore(usagePath)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := relay.New(store, usageStore, cfg)
	if err != nil {
		t.Fatal(err)
	}
	service := NewDesktopService(runtime)
	service.setNeedsOnboarding(true)
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatalf("deferred config exists before completion: %v", err)
	}
	if _, err := os.Stat(usagePath); !os.IsNotExist(err) {
		t.Fatalf("deferred usage exists before completion: %v", err)
	}
	if err := service.CompleteOnboarding(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config was not persisted after completion: %v", err)
	}
	if _, err := os.Stat(usagePath); err != nil {
		t.Fatalf("usage was not persisted after completion: %v", err)
	}
}

func TestSetProxyPortUpdatesRuntimeAndPersistsConfig(t *testing.T) {
	directory := t.TempDir()
	configPath := filepath.Join(directory, "config.json")
	store := config.NewStore(configPath)
	cfg := config.Default(18765)
	if err := store.Save(cfg); err != nil {
		t.Fatal(err)
	}
	runtime := newTestRuntime(t, directory, store, cfg)
	service := NewDesktopService(runtime)
	if err := service.SetProxyPort(18766); err != nil {
		t.Fatal(err)
	}
	if got := runtime.State().Config.ProxyPort; got != 18766 {
		t.Fatalf("runtime proxy port = %d", got)
	}
	persisted, err := store.LoadOrCreate(0)
	if err != nil {
		t.Fatal(err)
	}
	if persisted.ProxyPort != 18766 {
		t.Fatalf("persisted proxy port = %d", persisted.ProxyPort)
	}
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	occupiedPort := occupied.Addr().(*net.TCPAddr).Port
	if err := service.SetProxyPort(occupiedPort); err == nil {
		t.Fatal("occupied proxy port should fail")
	}
	if got := runtime.State().Config.ProxyPort; got != 18766 {
		t.Fatalf("failed port change should keep old port, got %d", got)
	}
	for _, port := range []int{0, 65536} {
		if err := service.SetProxyPort(port); err == nil {
			t.Fatalf("invalid proxy port %d should fail", port)
		}
	}
}

func TestDogeTokenSwitchPromptUsesAvailableTokensInCurrentCategoryAndDismissesForFiveMinutes(t *testing.T) {
	directory := t.TempDir()
	store := config.NewStore(filepath.Join(directory, "config.json"))
	cfg := config.Default(18765)
	cfg.Doge.AccessToken = "fake-doge-access-token"
	cfg.Doge.Groups = []string{"余额低价组", "余额稳定组"}
	cfg.Doge.Tokens = []config.DogeToken{
		{ID: 41, Status: 1, Name: "当前令牌", Category: config.CategoryCodex, Key: "sk-current", Group: "余额低价组", GroupDisplayName: "余额低价组", GroupRatio: 0.02},
		{ID: 42, Status: 1, Name: "Codex 低价组", Category: config.CategoryCodex, Key: "sk-candidate", Group: "余额低价组", GroupDisplayName: "GPT低价组", GroupRatio: 0.02},
		{ID: 44, Status: 2, Name: "已禁用令牌", Category: config.CategoryCodex, Key: "sk-disabled", Group: "余额低价组", GroupDisplayName: "余额低价组", GroupRatio: 0.02},
		{ID: 43, Status: 1, Name: "其他分组", Category: config.CategoryCodex, Key: "sk-other", Group: "余额稳定组", GroupDisplayName: "余额稳定组", GroupRatio: 0.02},
	}
	cfg.Profiles = []config.Profile{{
		ID: "doge-profile", Source: config.SourceDoge, Category: config.CategoryCodex, Name: "当前令牌",
		BaseURL: "https://example.test/v1", APIKey: "sk-current", RemoteTokenID: 41,
	}}
	cfg.ActiveProfiles = map[string]string{config.CategoryCodex: "doge-profile"}
	if err := store.Save(cfg); err != nil {
		t.Fatal(err)
	}
	runtime := newTestRuntime(t, directory, store, cfg)
	service := NewDesktopService(runtime)
	for index := 0; index < 5; index++ {
		runtime.ObserveUpstreamResult("doge-profile", config.CategoryCodex, 403, false)
	}
	prompt := service.GetState().Doge.TokenSwitch
	if prompt == nil || prompt.FailureKind != "auth" || prompt.FailureCount != 5 || len(prompt.Candidates) != 2 || prompt.Candidates[0].TokenID != 42 || prompt.Candidates[1].TokenID != 43 {
		t.Fatalf("token switch prompt = %+v", prompt)
	}
	if prompt.Candidates[0].Name != "Codex 低价组" || prompt.Candidates[0].Group != "GPT低价组" || prompt.Candidates[0].Ratio != 0.02 {
		t.Fatalf("candidate display data = %+v", prompt.Candidates[0])
	}
	if prompt.Candidates[1].Name != "其他分组" || prompt.Candidates[1].Group != "余额稳定组" || prompt.Candidates[1].Ratio != 0.02 {
		t.Fatalf("cross-group candidate display data = %+v", prompt.Candidates[1])
	}
	if err := service.DismissDogeTokenSwitch(prompt.Key); err != nil {
		t.Fatal(err)
	}
	if got := service.GetState().Doge.TokenSwitch; got != nil {
		t.Fatalf("dismissed prompt should be suppressed, got %+v", got)
	}
	service.switchMu.Lock()
	service.switchPrompts[prompt.Key].suppressedUntil = time.Now().Add(-time.Minute)
	service.switchMu.Unlock()
	if got := service.GetState().Doge.TokenSwitch; got != nil {
		t.Fatalf("ongoing failure should stay suppressed after five minutes, got %+v", got)
	}
}

func TestDogeDirectorySyncCreatesAndClearsTokenSwitchPrompt(t *testing.T) {
	previous := config.Default(18765)
	previous.Doge.AccessToken = "fake-doge-access-token"
	previous.Doge.Groups = []string{"余额低价组"}
	previous.Doge.Tokens = []config.DogeToken{
		{ID: 41, Status: 1, Name: "当前令牌", Category: config.CategoryCodex, Key: "sk-current", Group: "余额低价组", GroupDisplayName: "余额低价组", GroupRatio: 0.02},
		{ID: 42, Status: 1, Name: "备用令牌", Category: config.CategoryCodex, Key: "sk-candidate", Group: "余额低价组", GroupDisplayName: "余额低价组", GroupRatio: 0.02},
	}
	previous.Profiles = []config.Profile{{
		ID: "doge-profile", Source: config.SourceDoge, Category: config.CategoryCodex, Name: "当前令牌",
		BaseURL: "https://example.test/v1", APIKey: "sk-current", RemoteTokenID: 41,
	}}
	previous.ActiveProfiles = map[string]string{config.CategoryCodex: "doge-profile"}

	current := previous.Doge
	current.Tokens = []config.DogeToken{
		{ID: 42, Status: 1, Name: "备用令牌", Category: config.CategoryCodex, Key: "sk-candidate", Group: "余额低价组", GroupDisplayName: "余额低价组", GroupRatio: 0.02},
	}
	context := buildDogeDirectorySwitchContext(previous, current)
	if context == nil || context.failureKind != "directory" || context.directoryReason != dogeDirectoryFailureMissing || context.token.ID != 41 {
		t.Fatalf("missing token context = %+v", context)
	}
	prompt := publicDogeTokenSwitchPrompt(*context)
	if prompt.Message == "" || len(prompt.Candidates) != 1 || prompt.Candidates[0].TokenID != 42 {
		t.Fatalf("missing token prompt = %+v", prompt)
	}

	current.Tokens = []config.DogeToken{
		{ID: 41, Status: 2, Name: "当前令牌", Category: config.CategoryCodex, Key: "sk-current", Group: "余额低价组", GroupDisplayName: "余额低价组", GroupRatio: 0.02},
		{ID: 42, Status: 1, Name: "备用令牌", Category: config.CategoryCodex, Key: "sk-candidate", Group: "余额低价组", GroupDisplayName: "余额低价组", GroupRatio: 0.02},
	}
	context = buildDogeDirectorySwitchContext(previous, current)
	if context == nil || context.directoryReason != dogeDirectoryFailureUnavailable {
		t.Fatalf("unavailable token context = %+v", context)
	}
	if recovered := buildDogeDirectorySwitchContext(previous, previous.Doge); recovered != nil {
		t.Fatalf("recovered token should clear directory context = %+v", recovered)
	}
}

func newTestRuntime(t *testing.T, directory string, store *config.Store, cfg config.AppConfig) *relay.Runtime {
	t.Helper()
	usageStore, err := usage.NewStore(filepath.Join(directory, "usage.json"))
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := relay.New(store, usageStore, cfg)
	if err != nil {
		t.Fatal(err)
	}
	return runtime
}
