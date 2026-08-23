/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : 前端绑定服务与本地代理生命周期
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package desktop

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"codexrelay/internal/config"
	"codexrelay/internal/network"
	"codexrelay/internal/platform"
	"codexrelay/internal/relay"
	"codexrelay/internal/usage"

	"github.com/wailsapp/wails/v3/pkg/application"
)

var applicationVersion = "1.0.0"

const (
	NotificationKindBalance      = "balance"
	NotificationKindSubscription = "subscription"
	NotificationKindAnnouncement = "announcement"
	NotificationKindTokenSwitch  = "token-switch"
	dogeLowQuotaThresholdUSD     = 1.0
	dogeProfileURL               = "https://ergouzi.life/profile"
)

type DesktopState struct {
	Version          string                  `json:"version"`
	NeedsOnboarding  bool                    `json:"needsOnboarding"`
	DataDirectory    string                  `json:"dataDirectory"`
	ProxyPort        int                     `json:"proxyPort"`
	ProxyURL         string                  `json:"proxyUrl"`
	ProxyURLs        map[string]string       `json:"proxyUrls"`
	LocalAccessToken string                  `json:"localAccessToken"`
	ActiveProfiles   map[string]string       `json:"activeProfiles"`
	Profiles         []PublicProfile         `json:"profiles"`
	ClientConfigs    []PublicClientConfig    `json:"clientConfigs"`
	Network          network.Settings        `json:"network"`
	SystemProxy      network.SystemProxyInfo `json:"systemProxy"`
	Requests         []usage.RequestRecord   `json:"requests"`
	Usage            usage.Overview          `json:"usage"`
	UptimeSeconds    int64                   `json:"uptimeSeconds"`
	Preferences      config.Preferences      `json:"preferences"`
	Doge             DogeState               `json:"doge"`
}

type DogeState struct {
	BaseURL             string                       `json:"baseUrl"`
	Bound               bool                         `json:"bound"`
	Account             PublicDogeAccount            `json:"account"`
	User                map[string]any               `json:"user"`
	WalletUSD           float64                      `json:"walletUsd"`
	SubscriptionsUSD    float64                      `json:"subscriptionsUsd"`
	TotalUSD            float64                      `json:"totalUsd"`
	Subscriptions       []PublicDogeSubscription     `json:"subscriptions"`
	RedemptionEnabled   bool                         `json:"redemptionEnabled"`
	TopupLink           string                       `json:"topupLink"`
	Groups              []string                     `json:"groups"`
	Tokens              []PublicDogeToken            `json:"tokens"`
	Notifications       PublicDogeNotifications      `json:"notifications"`
	TokenSwitch         *PublicDogeTokenSwitchPrompt `json:"tokenSwitch,omitempty"`
	Syncing             bool                         `json:"syncing"`
	SyncPhase           string                       `json:"syncPhase"`
	AnnouncementSyncing bool                         `json:"announcementSyncing"`
	SyncIntervalMinutes int                          `json:"syncIntervalMinutes"`
	LastSyncAt          string                       `json:"lastSyncAt"`
	LastSyncError       string                       `json:"lastSyncError"`
}

// PublicDogeAccount 是连接页展示的账户摘要；额度字段统一转换为美元，不包含访问令牌。
type PublicDogeAccount struct {
	UserID       int64   `json:"userId"`
	Nickname     string  `json:"nickname"`
	Email        string  `json:"email"`
	BalanceUSD   float64 `json:"balanceUsd"`
	UsedUSD      float64 `json:"usedUsd"`
	RequestCount int64   `json:"requestCount"`
}

// PublicDogeAnnouncement 是公告面板使用的公开公告内容，不包含访问令牌。
type PublicDogeAnnouncement struct {
	ID          int64  `json:"id"`
	Content     string `json:"content"`
	Extra       string `json:"extra"`
	PublishDate string `json:"publishDate"`
	Type        string `json:"type"`
	Read        bool   `json:"read"`
}

// PublicDogeAlert 是右下角提醒窗使用的单条提醒摘要。
type PublicDogeAlert struct {
	Kind           string  `json:"kind"`
	Key            string  `json:"key"`
	Title          string  `json:"title"`
	Message        string  `json:"message"`
	AnnouncementID int64   `json:"announcementId,omitempty"`
	AmountUSD      float64 `json:"amountUsd,omitempty"`
}

// PublicDogeTokenSwitchCandidate 是令牌切换弹窗展示的候选摘要，不包含完整密钥；Group 始终是 display_name。
type PublicDogeTokenSwitchCandidate struct {
	TokenID        int64   `json:"tokenId"`
	ProfileID      string  `json:"profileId,omitempty"`
	Name           string  `json:"name"`
	Source         string  `json:"source"`
	Group          string  `json:"group"`
	Ratio          float64 `json:"ratio"`
	Selectable     bool    `json:"selectable"`
	DisabledReason string  `json:"disabledReason,omitempty"`
}

// PublicDogeTokenSwitchPrompt 是右下角令牌切换弹窗的只读状态。
type PublicDogeTokenSwitchPrompt struct {
	Key            string                           `json:"key"`
	Category       string                           `json:"category"`
	FailureKind    string                           `json:"failureKind"`
	FailureCount   int                              `json:"failureCount"`
	FailureStatus  int                              `json:"failureStatus,omitempty"`
	CurrentTokenID int64                            `json:"currentTokenId"`
	CurrentName    string                           `json:"currentName"`
	CurrentGroup   string                           `json:"currentGroup"`
	CurrentRatio   float64                          `json:"currentRatio"`
	Message        string                           `json:"message"`
	Candidates     []PublicDogeTokenSwitchCandidate `json:"candidates"`
}

// PublicDogeNotifications 是公告中心和提醒窗口共用的只读快照。
type PublicDogeNotifications struct {
	Initialized   bool                     `json:"initialized"`
	Enabled       bool                     `json:"enabled"`
	CurrentNotice string                   `json:"currentNotice"`
	Announcements []PublicDogeAnnouncement `json:"announcements"`
	UnreadCount   int                      `json:"unreadCount"`
	Alerts        []PublicDogeAlert        `json:"alerts"`
	LastSyncAt    string                   `json:"lastSyncAt"`
	LastSyncError string                   `json:"lastSyncError"`
	Syncing       bool                     `json:"syncing"`
}

// PublicDogeSubscription 是提供给界面的套餐剩余摘要，金额统一转换为美元。
type PublicDogeSubscription struct {
	ID           int64   `json:"id"`
	PlanID       int64   `json:"planId"`
	PlanTitle    string  `json:"planTitle"`
	Status       string  `json:"status"`
	RemainingUSD float64 `json:"remainingUsd"`
	EndTime      int64   `json:"endTime"`
}

// PublicDogeToken 是主页令牌目录的只读摘要；Group 保留原始分组键供内部状态使用，界面只读取 GroupDisplayName。
// Permitted 表示令牌状态、分组权限和本地完整密钥均满足切换约束，不代表已经导入 Profile。
type PublicDogeToken struct {
	ID               int64   `json:"id"`
	MaskedKey        string  `json:"maskedKey"`
	OrderKey         string  `json:"orderKey"`
	Status           int     `json:"status"`
	Name             string  `json:"name"`
	CreatedTime      int64   `json:"createdTime"`
	AccessedTime     int64   `json:"accessedTime"`
	ExpiredTime      int64   `json:"expiredTime"`
	RemainQuota      int64   `json:"remainQuota"`
	UnlimitedQuota   bool    `json:"unlimitedQuota"`
	UsedQuota        int64   `json:"usedQuota"`
	Group            string  `json:"group"`
	GroupDisplayName string  `json:"groupDisplayName"`
	GroupRatio       float64 `json:"groupRatio"`
	Category         string  `json:"category"`
	Note             string  `json:"note"`
	NeedsCategory    bool    `json:"needsCategory"`
	Permitted        bool    `json:"permitted"`
	Imported         bool    `json:"imported"`
	ProfileID        string  `json:"profileId"`
	Active           bool    `json:"active"`
}

// DogeTokenCategoryInput 描述一个二狗子令牌的本地存放类别选择；ID 必须来自当前同步目录。
type DogeTokenCategoryInput struct {
	ID       int64  `json:"id"`
	Category string `json:"category"`
}

type PublicProfile struct {
	ID            string            `json:"id"`
	Source        string            `json:"source"`
	Category      string            `json:"category"`
	Name          string            `json:"name"`
	BaseURL       string            `json:"baseUrl"`
	APIKey        string            `json:"apiKey"`
	Note          string            `json:"note,omitempty"`
	Headers       map[string]string `json:"headers"`
	Models        []PublicModel     `json:"models"`
	DefaultModel  string            `json:"defaultModel"`
	Active        bool              `json:"active"`
	PreviewURL    string            `json:"previewUrl"`
	RemoteTokenID int64             `json:"remoteTokenId,omitempty"`
}

// PublicModel 是编辑页模型管理器展示的模型摘要，不包含密钥或请求正文。
type PublicModel struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	OwnedBy       string `json:"ownedBy"`
	ContextWindow int64  `json:"contextWindow,omitempty"`
}

// ModelInput 是前端保存模型目录时提交的单条模型；ID 必须是用户确认的真实模型 ID。
type ModelInput struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	OwnedBy       string `json:"ownedBy"`
	ContextWindow int64  `json:"contextWindow"`
}

type ProfileInput struct {
	ID           string            `json:"id"`
	Source       string            `json:"source"`
	Category     string            `json:"category"`
	Name         string            `json:"name"`
	BaseURL      string            `json:"baseUrl"`
	APIKey       string            `json:"apiKey"`
	Note         string            `json:"note"`
	Headers      map[string]string `json:"headers"`
	Models       []ModelInput      `json:"models"`
	DefaultModel string            `json:"defaultModel"`
}

type TestResult struct {
	OK         bool   `json:"ok"`
	Reachable  bool   `json:"reachable"`
	Status     int    `json:"status"`
	DurationMs int64  `json:"durationMs"`
	URL        string `json:"url"`
}

type DesktopService struct {
	runtime             *relay.Runtime
	server              *http.Server
	listener            net.Listener
	mu                  sync.Mutex
	onStateChanged      func()
	stateChanged        []func()
	needsOnboarding     bool
	dogeMu              sync.Mutex
	dogeSyncing         bool
	dogeSyncPhase       string
	announcementSyncing bool
	switchMu            sync.Mutex
	switchPrompts       map[string]*tokenSwitchPromptState
	directorySwitch     *tokenSwitchContext
	syncCancel          context.CancelFunc
	syncStarted         bool
}

type tokenSwitchPromptState struct {
	dismissed       bool
	suppressedUntil time.Time
}

type tokenSwitchContext struct {
	key             string
	failureKind     string
	directoryReason string
	failureCount    int
	failureStatus   int
	health          relay.HealthSnapshot
	profile         config.Profile
	token           config.DogeToken
	candidates      []config.DogeToken
	profilesByID    map[int64]config.Profile
	tokens          []config.DogeToken
	groups          []string
}

const (
	dogeDirectoryFailureMissing     = "missing"
	dogeDirectoryFailureUnavailable = "unavailable"
)

func NewDesktopService(runtime *relay.Runtime) *DesktopService {
	service := &DesktopService{runtime: runtime, switchPrompts: make(map[string]*tokenSwitchPromptState)}
	runtime.SetHealthChangedHandler(service.notifyStateChanged)
	return service
}

func (s *DesktopService) setStateChangedHandler(handler func()) {
	s.mu.Lock()
	s.onStateChanged = handler
	s.mu.Unlock()
}

func (s *DesktopService) addStateChangedHandler(handler func()) {
	if handler == nil {
		return
	}
	s.mu.Lock()
	s.stateChanged = append(s.stateChanged, handler)
	s.mu.Unlock()
}

func (s *DesktopService) notifyStateChanged() {
	s.mu.Lock()
	handler := s.onStateChanged
	handlers := append([]func(){}, s.stateChanged...)
	s.mu.Unlock()
	if handler != nil {
		handler()
	}
	for _, callback := range handlers {
		callback()
	}
}

// setNeedsOnboarding 设置本次进程的首次启动引导状态；状态不写入配置，配置与用量文件存在后下次启动自然跳过。
func (s *DesktopService) setNeedsOnboarding(value bool) {
	s.mu.Lock()
	changed := s.needsOnboarding != value
	s.needsOnboarding = value
	s.mu.Unlock()
	if changed {
		s.notifyStateChanged()
	}
}

func (s *DesktopService) onboardingStatus() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.needsOnboarding
}

// setDogeSyncing 更新账户或公告同步状态，并立即通知前端；状态只存在运行时，不写入配置文件。
func (s *DesktopService) setDogeSyncing(announcement bool, syncing bool) {
	s.mu.Lock()
	if announcement {
		s.announcementSyncing = syncing
	} else {
		s.dogeSyncing = syncing
	}
	s.mu.Unlock()
	s.notifyStateChanged()
}

func (s *DesktopService) setDogeSyncPhase(phase string) {
	s.mu.Lock()
	s.dogeSyncPhase = phase
	s.mu.Unlock()
	s.notifyStateChanged()
}

func (s *DesktopService) dogeSyncStatus() (bool, string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dogeSyncing, s.dogeSyncPhase, s.announcementSyncing
}

func (s *DesktopService) updateConfig(mutator func(*config.AppConfig) error) error {
	_, err := s.runtime.UpdateConfig(mutator)
	if err == nil {
		s.notifyStateChanged()
	}
	return err
}

func (s *DesktopService) ServiceStartup(_ context.Context, _ application.ServiceOptions) error {
	state := s.runtime.State()
	if state == nil {
		return errors.New("代理尚未初始化")
	}
	// 启动时只探测各客户端的已知配置目录；延迟存储会在首次初始化完成后再落盘。
	if err := s.scanClientConfigs(); err != nil {
		application.Get().Logger.Warn("扫描外部客户端配置目录失败", "error", err)
	}
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(state.Config.ProxyPort))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("代理端口 %s 无法监听，可能已有实例正在运行: %w", address, err)
	}
	server := &http.Server{
		Handler: s.runtime.ProxyHandler(), ReadHeaderTimeout: 15 * time.Second,
		IdleTimeout: 120 * time.Second,
	}
	s.mu.Lock()
	s.listener = listener
	s.server = server
	s.mu.Unlock()
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			application.Get().Logger.Error("透明代理异常退出", "error", err)
		}
	}()
	s.mu.Lock()
	if !s.syncStarted {
		s.syncStarted = true
		ctx, cancel := context.WithCancel(context.Background())
		s.syncCancel = cancel
		s.mu.Unlock()
		go s.dogeSyncLoop(ctx)
		if strings.TrimSpace(state.Config.Doge.AccessToken) != "" {
			// 启动恢复只刷新目录元数据；已有密钥来自本地缓存，新令牌才按需补全。
			go func() { _ = s.syncDoge(context.Background(), "", false, dogeSyncMetadata) }()
		} else {
			go func() { _ = s.SyncDogeAnnouncements() }()
		}
	} else {
		s.mu.Unlock()
	}
	return nil
}

func (s *DesktopService) ServiceShutdown() error {
	s.mu.Lock()
	server := s.server
	cancel := s.syncCancel
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}

func (s *DesktopService) GetState() DesktopState {
	state := s.runtime.State()
	if state == nil {
		return DesktopState{}
	}
	dogeSyncing, dogeSyncPhase, announcementSyncing := s.dogeSyncStatus()
	profiles := make([]PublicProfile, 0, len(state.Config.Profiles))
	for _, profile := range state.Config.Profiles {
		profiles = append(profiles, publicProfile(profile, state.Config.ActiveProfiles))
	}
	profilesByRemoteID := make(map[int64]PublicProfile)
	for _, profile := range profiles {
		if profile.RemoteTokenID > 0 {
			profilesByRemoteID[profile.RemoteTokenID] = profile
		}
	}
	dogeTokens := make([]PublicDogeToken, 0, len(state.Config.Doge.Tokens))
	for _, token := range state.Config.Doge.Tokens {
		imported := profilesByRemoteID[token.ID]
		category := token.Category
		if category == "" && imported.ID != "" {
			category = imported.Category
		}
		masked := token.MaskedKey
		if masked == "" {
			masked = maskDogeKey(token.Key)
		}
		masked = normalizeDogeAPIKey(masked)
		note := token.Note
		if imported.Note != "" && !strings.HasPrefix(imported.Note, "二狗子 · 分组：") {
			note = imported.Note
		}
		if note == "" {
			note = dogeTokenNote(token)
		}
		dogeTokens = append(dogeTokens, PublicDogeToken{
			ID: token.ID, MaskedKey: masked, OrderKey: dogeTokenOrderKey(token), Status: token.Status, Name: token.Name,
			CreatedTime: token.CreatedTime, AccessedTime: token.AccessedTime, ExpiredTime: token.ExpiredTime,
			RemainQuota: token.RemainQuota, UnlimitedQuota: token.UnlimitedQuota, UsedQuota: token.UsedQuota,
			Group: token.Group, GroupDisplayName: token.GroupDisplayName, GroupRatio: token.GroupRatio,
			Category: category, Note: note, NeedsCategory: category == "", Permitted: dogeTokenSwitchable(token, state.Config.Doge.Groups),
			Imported: imported.ID != "", ProfileID: imported.ID, Active: imported.Active,
		})
	}
	proxyURLs := make(map[string]string, len(config.Categories))
	for _, category := range config.Categories {
		proxyURLs[category] = fmt.Sprintf("http://127.0.0.1:%d/%s", state.Config.ProxyPort, category)
	}
	publicSubscriptions := make([]PublicDogeSubscription, 0, len(state.Config.Doge.Subscriptions))
	subscriptionsUSD := 0.0
	for _, subscription := range state.Config.Doge.Subscriptions {
		remaining := subscription.AmountTotal - subscription.AmountUsed
		remainingUSD := dogeQuotaToUSD(remaining)
		subscriptionsUSD += remainingUSD
		publicSubscriptions = append(publicSubscriptions, PublicDogeSubscription{ID: subscription.ID, PlanID: subscription.PlanID, PlanTitle: subscription.PlanTitle, Status: subscription.Status, RemainingUSD: remainingUSD, EndTime: subscription.EndTime})
	}
	walletUSD := dogeQuotaToUSD(state.Config.Doge.Account.Quota)
	account := PublicDogeAccount{
		UserID: state.Config.Doge.Account.ID, Nickname: state.Config.Doge.Account.DisplayName,
		Email: state.Config.Doge.Account.Email, BalanceUSD: walletUSD,
		UsedUSD: dogeQuotaToUSD(state.Config.Doge.Account.UsedQuota), RequestCount: state.Config.Doge.Account.RequestCount,
	}
	dogeState := DogeState{
		BaseURL: state.Config.Doge.BaseURL, Bound: strings.TrimSpace(state.Config.Doge.AccessToken) != "", Account: account,
		WalletUSD: walletUSD, SubscriptionsUSD: subscriptionsUSD, TotalUSD: walletUSD + subscriptionsUSD,
		Subscriptions: publicSubscriptions, RedemptionEnabled: state.Config.Doge.Topup.EnableRedemption, TopupLink: state.Config.Doge.Topup.TopupLink,
		User: state.Config.Doge.User, Groups: append([]string(nil), state.Config.Doge.Groups...), Tokens: dogeTokens,
		Notifications:       publicDogeNotifications(state.Config.Doge, walletUSD, publicSubscriptions, announcementSyncing),
		Syncing:             dogeSyncing,
		SyncPhase:           dogeSyncPhase,
		AnnouncementSyncing: announcementSyncing,
		SyncIntervalMinutes: state.Config.Doge.SyncIntervalMinutes, LastSyncError: state.Config.Doge.LastSyncError,
	}
	dogeState.TokenSwitch = s.buildDogeTokenSwitchPrompt()
	if !state.Config.Doge.LastSyncAt.IsZero() {
		dogeState.LastSyncAt = state.Config.Doge.LastSyncAt.Format(time.RFC3339)
	}
	return DesktopState{
		Version: applicationVersion, NeedsOnboarding: s.onboardingStatus(), DataDirectory: s.runtime.DataDirectory(), ProxyPort: state.Config.ProxyPort,
		ProxyURL: proxyURLs[config.CategoryCodex], ProxyURLs: proxyURLs,
		LocalAccessToken: state.Config.LocalAccessToken, ActiveProfiles: state.Config.ActiveProfiles,
		Profiles: profiles, ClientConfigs: publicClientConfigs(state.Config), Network: state.Config.Network, SystemProxy: state.SystemProxy,
		Requests: s.runtime.RecentRecords(), Usage: s.runtime.UsageOverview(), UptimeSeconds: int64(s.runtime.Uptime().Seconds()),
		Preferences: state.Config.Preferences, Doge: dogeState,
	}
}

// CompleteOnboarding 结束首次启动引导并启用便携数据持久化；跳过和绑定成功都调用此方法。
func (s *DesktopService) CompleteOnboarding() error {
	if err := s.runtime.ActivatePortablePersistence(); err != nil {
		return fmt.Errorf("保存首次初始化数据: %w", err)
	}
	s.setNeedsOnboarding(false)
	return nil
}

// buildDogeTokenSwitchPrompt 根据运行时健康快照构建当前唯一的令牌切换提示。
// 只观察活动的二狗子 Profile；候选令牌复用主窗口类别下的可用令牌集合。
// 用户取消后的抑制状态只保存在内存中，失败状态恢复后会被清理，下一次新的阈值事件可以再次提示。
func (s *DesktopService) buildDogeTokenSwitchPrompt() *PublicDogeTokenSwitchPrompt {
	state := s.runtime.State()
	if state == nil || strings.TrimSpace(state.Config.Doge.AccessToken) == "" {
		return nil
	}

	profilesByRemoteID := make(map[int64]config.Profile)
	for _, profile := range state.Config.Profiles {
		if profile.Source == config.SourceDoge && profile.RemoteTokenID > 0 {
			profilesByRemoteID[profile.RemoteTokenID] = profile
		}
	}
	byID := make(map[int64]config.DogeToken, len(state.Config.Doge.Tokens))
	for _, token := range state.Config.Doge.Tokens {
		byID[token.ID] = token
	}

	snapshots := s.runtime.HealthSnapshots()
	type triggeredContext struct {
		context  tokenSwitchContext
		priority int
	}
	contexts := make([]triggeredContext, 0, len(snapshots))
	activeKeys := make(map[string]struct{}, len(snapshots))
	if directoryContext := s.dogeDirectorySwitchContext(); directoryContext != nil &&
		state.Config.ActiveProfiles[directoryContext.profile.Category] == directoryContext.profile.ID {
		activeKeys[directoryContext.key] = struct{}{}
		contexts = append(contexts, triggeredContext{context: *directoryContext, priority: -1})
	}
	for _, health := range snapshots {
		failureKind, failureCount, priority := "", 0, 0
		switch {
		case health.AuthTriggered:
			failureKind, failureCount, priority = "auth", health.AuthFailures, 0
		case health.UpstreamTriggered:
			failureKind, failureCount, priority = "upstream", health.UpstreamFailures, 1
		default:
			continue
		}
		profileIndex := config.FindProfileIndex(state.Config.Profiles, health.ProfileID)
		if profileIndex < 0 {
			continue
		}
		profile := state.Config.Profiles[profileIndex]
		if profile.Source != config.SourceDoge || profile.RemoteTokenID <= 0 || state.Config.ActiveProfiles[profile.Category] != profile.ID {
			continue
		}
		current, ok := byID[profile.RemoteTokenID]
		if !ok {
			continue
		}
		key := profile.ID + "|" + failureKind + "|" + strconv.FormatUint(health.TriggerGeneration, 10)
		activeKeys[key] = struct{}{}
		ctx := tokenSwitchContext{
			key: key, failureKind: failureKind, failureCount: failureCount, failureStatus: health.LastStatus,
			health: health, profile: profile, token: current, profilesByID: profilesByRemoteID,
			tokens: state.Config.Doge.Tokens, groups: state.Config.Doge.Groups,
		}
		contexts = append(contexts, triggeredContext{context: ctx, priority: priority})
	}

	switchPromptStates := s.switchPromptStates(activeKeys)
	sort.Slice(contexts, func(i, j int) bool {
		if contexts[i].priority != contexts[j].priority {
			return contexts[i].priority < contexts[j].priority
		}
		return contexts[i].context.key < contexts[j].context.key
	})
	for _, item := range contexts {
		if dismissed, ok := switchPromptStates[item.context.key]; ok && dismissed {
			continue
		}
		return publicDogeTokenSwitchPrompt(item.context)
	}
	return nil
}

// dogeDirectorySwitchContext 返回同步目录检测生成的运行时提示快照。
// 快照保留同步前当前令牌的信息，使令牌从新目录消失后仍能按当前类别筛选候选；不写入配置文件。
func (s *DesktopService) dogeDirectorySwitchContext() *tokenSwitchContext {
	s.switchMu.Lock()
	defer s.switchMu.Unlock()
	if s.directorySwitch == nil {
		return nil
	}
	context := *s.directorySwitch
	context.tokens = append([]config.DogeToken(nil), s.directorySwitch.tokens...)
	context.groups = append([]string(nil), s.directorySwitch.groups...)
	context.profilesByID = make(map[int64]config.Profile, len(s.directorySwitch.profilesByID))
	for id, profile := range s.directorySwitch.profilesByID {
		context.profilesByID[id] = profile
	}
	return &context
}

// setDogeDirectorySwitchContext 替换自动同步检测到的目录失效提示。
// 提示 key 在同一活动令牌持续失效期间保持稳定，用户取消后沿用现有五分钟及持续期间抑制规则。
func (s *DesktopService) setDogeDirectorySwitchContext(context *tokenSwitchContext) {
	s.switchMu.Lock()
	previousKey := ""
	if s.directorySwitch != nil {
		previousKey = s.directorySwitch.key
	}
	nextKey := ""
	if context != nil {
		nextKey = context.key
	}
	changed := previousKey != nextKey
	s.directorySwitch = context
	s.switchMu.Unlock()
	if changed || context != nil {
		s.notifyStateChanged()
	}
}

// switchPromptStates 同步清理已经不再触发的提示状态，并返回当前仍允许展示的状态。
func (s *DesktopService) switchPromptStates(activeKeys map[string]struct{}) map[string]bool {
	now := time.Now()
	s.switchMu.Lock()
	defer s.switchMu.Unlock()
	visible := make(map[string]bool, len(activeKeys))
	suppressedBases := make(map[string]struct{})
	for key := range s.switchPrompts {
		promptState := s.switchPrompts[key]
		_, active := activeKeys[key]
		if !active {
			// 失败状态刚恢复时保留五分钟抑制窗口，避免短暂恢复后再次连续失败立即弹窗。
			if promptState.dismissed && now.Before(promptState.suppressedUntil) {
				suppressedBases[switchPromptBaseKey(key)] = struct{}{}
				continue
			}
			if !promptState.dismissed || !now.Before(promptState.suppressedUntil) {
				delete(s.switchPrompts, key)
			}
			continue
		}
		if promptState.dismissed {
			// 用户取消后，只要同一失败状态仍在持续，就不因五分钟到期再次打扰。
			visible[key] = true
		}
	}
	for key := range activeKeys {
		if _, ok := s.switchPrompts[key]; !ok {
			s.switchPrompts[key] = &tokenSwitchPromptState{}
		}
		if _, suppressed := suppressedBases[switchPromptBaseKey(key)]; suppressed {
			visible[key] = true
		}
	}
	return visible
}

func switchPromptBaseKey(key string) string {
	if index := strings.LastIndex(key, "|"); index >= 0 {
		return key[:index]
	}
	return key
}

// publicDogeTokenSwitchPrompt 将内部令牌和 Profile 转为不含完整密钥的前端提示。
func publicDogeTokenSwitchPrompt(context tokenSwitchContext) *PublicDogeTokenSwitchPrompt {
	currentName := strings.TrimSpace(context.profile.Name)
	if currentName == "" {
		currentName = context.token.Name
	}
	candidates := make([]PublicDogeTokenSwitchCandidate, 0)
	for _, token := range availableDogeTokensForCategory(context.tokens, context.profilesByID, context.groups, context.profile.Category) {
		if token.ID == context.token.ID {
			continue
		}
		category := token.Category
		profile, imported := context.profilesByID[token.ID]
		if !config.IsCategory(category) {
			continue
		}
		name := strings.TrimSpace(token.Name)
		if imported && strings.TrimSpace(profile.Name) != "" {
			name = strings.TrimSpace(profile.Name)
		}
		if name == "" {
			name = "二狗子令牌 " + strconv.FormatInt(token.ID, 10)
		}
		group := dogeTokenDisplayGroup(token)
		candidates = append(candidates, PublicDogeTokenSwitchCandidate{
			TokenID: token.ID, ProfileID: profile.ID, Name: name, Source: "二狗子 API",
			Group: group, Ratio: token.GroupRatio, Selectable: true,
		})
	}
	failureMessage := fmt.Sprintf("当前令牌“%s”在 3 分钟内出现 %d 次上游异常。上游连续异常，是否尝试切换同类别的其他可用令牌？", currentName, context.failureCount)
	if context.failureKind == "auth" {
		failureMessage = fmt.Sprintf("当前令牌“%s”连续 %d 次返回 HTTP %d，是否切换同类别的其他可用令牌？", currentName, context.failureCount, context.failureStatus)
	} else if context.failureKind == "directory" {
		if context.directoryReason == dogeDirectoryFailureMissing {
			failureMessage = fmt.Sprintf("当前令牌“%s”已从最新令牌目录中消失，是否切换同类别的其他可用令牌？", currentName)
		} else {
			failureMessage = fmt.Sprintf("当前令牌“%s”在最新令牌目录中已失效，是否切换同类别的其他可用令牌？", currentName)
		}
	}
	currentGroup := dogeTokenDisplayGroup(context.token)
	return &PublicDogeTokenSwitchPrompt{
		Key: context.key, Category: context.profile.Category, FailureKind: context.failureKind,
		FailureCount: context.failureCount, FailureStatus: context.failureStatus, CurrentTokenID: context.token.ID,
		CurrentName: currentName, CurrentGroup: currentGroup, CurrentRatio: context.token.GroupRatio,
		Message: failureMessage, Candidates: candidates,
	}
}

func publicDogeNotifications(connection config.DogeConnection, walletUSD float64, subscriptions []PublicDogeSubscription, syncing bool) PublicDogeNotifications {
	readIDs := make(map[int64]struct{}, len(connection.Notifications.ReadAnnouncementIDs))
	for _, id := range connection.Notifications.ReadAnnouncementIDs {
		readIDs[id] = struct{}{}
	}
	dismissed := make(map[string]struct{}, len(connection.Notifications.DismissedAlertKeys))
	for _, key := range connection.Notifications.DismissedAlertKeys {
		dismissed[key] = struct{}{}
	}
	publicAnnouncements := make([]PublicDogeAnnouncement, 0, len(connection.Notifications.Announcements))
	unread := 0
	for _, announcement := range connection.Notifications.Announcements {
		_, read := readIDs[announcement.ID]
		if !read {
			unread++
		}
		publicAnnouncements = append(publicAnnouncements, PublicDogeAnnouncement{
			ID: announcement.ID, Content: announcement.Content, Extra: announcement.Extra,
			PublishDate: announcement.PublishDate, Type: announcement.Type, Read: read,
		})
	}
	alerts := make([]PublicDogeAlert, 0)
	if connection.Notifications.Initialized && connection.Notifications.AnnouncementsEnabled {
		for _, announcement := range connection.Notifications.Announcements {
			key := announcementAlertKey(announcement.ID)
			if _, read := readIDs[announcement.ID]; read {
				continue
			}
			if _, ok := dismissed[key]; ok {
				continue
			}
			alerts = append(alerts, PublicDogeAlert{Kind: NotificationKindAnnouncement, Key: key, Title: "新的系统公告", Message: "平台发布了新的公告", AnnouncementID: announcement.ID})
		}
	}
	if connection.Account.ID > 0 && walletUSD <= dogeLowQuotaThresholdUSD {
		key := balanceAlertKey(connection.Account.ID)
		if _, ok := dismissed[key]; !ok {
			alerts = append(alerts, PublicDogeAlert{Kind: NotificationKindBalance, Key: key, Title: "余额提醒", Message: fmt.Sprintf("钱包余额仅剩 %s", formatDogeUSDValue(walletUSD)), AmountUSD: walletUSD})
		}
	}
	for _, subscription := range subscriptions {
		if subscription.Status != "active" || subscription.RemainingUSD > dogeLowQuotaThresholdUSD {
			continue
		}
		key := subscriptionAlertKey(subscription.ID)
		if _, ok := dismissed[key]; ok {
			continue
		}
		label := subscription.PlanTitle
		if strings.TrimSpace(label) == "" {
			label = fmt.Sprintf("套餐 %d", subscription.PlanID)
		}
		alerts = append(alerts, PublicDogeAlert{Kind: NotificationKindSubscription, Key: key, Title: "套餐提醒", Message: fmt.Sprintf("%s 剩余 %s", label, formatDogeUSDValue(subscription.RemainingUSD)), AmountUSD: subscription.RemainingUSD})
	}
	lastSyncAt := ""
	if !connection.Notifications.LastAnnouncementSyncAt.IsZero() {
		lastSyncAt = connection.Notifications.LastAnnouncementSyncAt.Format(time.RFC3339)
	}
	return PublicDogeNotifications{
		Initialized:   connection.Notifications.Initialized,
		Enabled:       connection.Notifications.AnnouncementsEnabled,
		CurrentNotice: connection.Notifications.CurrentNotice,
		Announcements: publicAnnouncements,
		UnreadCount:   unread,
		Alerts:        alerts,
		LastSyncAt:    lastSyncAt,
		LastSyncError: connection.Notifications.LastAnnouncementSyncError,
		Syncing:       syncing,
	}
}

// dogeQuotaToUSD 使用二狗子当前实例的额度换算规则将原始 quota 转为美元。
// 该规则来自当前接口样本：500000 quota = 1 美元；配置仍保留原始整数额度。
func dogeQuotaToUSD(quota int64) float64 {
	return float64(quota) / 500000
}

func formatDogeUSDValue(value float64) string {
	return fmt.Sprintf("$%.2f", value)
}

func balanceAlertKey(userID int64) string {
	return fmt.Sprintf("balance:%d", userID)
}

func subscriptionAlertKey(subscriptionID int64) string {
	return fmt.Sprintf("subscription:%d", subscriptionID)
}

func announcementAlertKey(announcementID int64) string {
	return fmt.Sprintf("announcement:%d", announcementID)
}

func (s *DesktopService) ReorderProfiles(ids []string) error {
	return s.updateConfig(func(cfg *config.AppConfig) error {
		ordered, err := config.OrderProfiles(cfg.Profiles, ids)
		if err != nil {
			return err
		}
		cfg.Profiles = ordered
		return nil
	})
}

// ReorderDogeTokens 按二狗子远端令牌 ID 保存主页顺序；名称、分组和 API 密钥只用于展示或访问，
// 不参与令牌身份定位。
func (s *DesktopService) ReorderDogeTokens(orderKeys []string) error {
	return s.updateConfig(func(cfg *config.AppConfig) error {
		known := make(map[string]struct{}, len(cfg.Doge.Tokens))
		for _, token := range cfg.Doge.Tokens {
			key := dogeTokenOrderKey(token)
			if key == "" {
				continue
			}
			known[key] = struct{}{}
		}
		seen := make(map[string]struct{}, len(orderKeys))
		next := make([]string, 0, len(known))
		for _, key := range orderKeys {
			key = strings.TrimSpace(key)
			if key == "" {
				return errors.New("二狗子令牌排序 ID 不能为空")
			}
			if _, ok := known[key]; !ok {
				return errors.New("二狗子令牌排序包含未知 ID")
			}
			if _, ok := seen[key]; ok {
				return errors.New("二狗子令牌排序包含重复 ID")
			}
			seen[key] = struct{}{}
			next = append(next, key)
		}
		for _, token := range cfg.Doge.Tokens {
			key := dogeTokenOrderKey(token)
			if key == "" {
				continue
			}
			if _, ok := seen[key]; !ok {
				next = append(next, key)
			}
		}
		cfg.Doge.TokenOrder = next
		cfg.Doge.Tokens = orderDogeTokens(next, cfg.Doge.Tokens)
		return nil
	})
}

// SetDogeTokenCategories 保存用户为尚未导入令牌选择的本地 API 类别；远端 group 仍用于权限判断，display_name 只作为界面文案。
// 已经导入的令牌不能通过该入口改类别，避免绕过编辑页导致启用映射与配置不一致。
func (s *DesktopService) SetDogeTokenCategories(assignments []DogeTokenCategoryInput) error {
	if len(assignments) == 0 {
		return errors.New("至少选择一个二狗子令牌类别")
	}
	return s.updateConfig(func(cfg *config.AppConfig) error {
		byID := make(map[int64]int, len(cfg.Doge.Tokens))
		for index := range cfg.Doge.Tokens {
			byID[cfg.Doge.Tokens[index].ID] = index
		}
		seen := make(map[int64]struct{}, len(assignments))
		for _, assignment := range assignments {
			if assignment.ID <= 0 {
				return errors.New("二狗子令牌 ID 无效")
			}
			if _, ok := seen[assignment.ID]; ok {
				return fmt.Errorf("二狗子令牌 %d 重复选择类别", assignment.ID)
			}
			seen[assignment.ID] = struct{}{}
			if !config.IsCategory(assignment.Category) {
				return fmt.Errorf("二狗子令牌 %d 的存放类别无效", assignment.ID)
			}
			index, ok := byID[assignment.ID]
			if !ok {
				return fmt.Errorf("二狗子令牌 %d 不存在，请先刷新目录", assignment.ID)
			}
			for _, profile := range cfg.Profiles {
				if profile.Source == config.SourceDoge && profile.RemoteTokenID == assignment.ID {
					return fmt.Errorf("二狗子令牌 %d 已导入，不能通过同步弹窗修改类别", assignment.ID)
				}
			}
			cfg.Doge.Tokens[index].Category = assignment.Category
		}
		return nil
	})
}

func (s *DesktopService) ClearUsage() error {
	if err := s.runtime.ClearUsage(); err != nil {
		return err
	}
	s.notifyStateChanged()
	return nil
}

func (s *DesktopService) SaveProfile(input ProfileInput) error {
	input.Name = strings.TrimSpace(input.Name)
	input.BaseURL = strings.TrimSpace(input.BaseURL)
	input.APIKey = strings.TrimSpace(input.APIKey)
	input.Note = strings.TrimSpace(input.Note)
	input.DefaultModel = strings.TrimSpace(input.DefaultModel)
	if input.Headers == nil {
		input.Headers = map[string]string{}
	}
	modelsOmitted := input.Models == nil
	return s.updateConfig(func(cfg *config.AppConfig) error {
		index := config.FindProfileIndex(cfg.Profiles, input.ID)
		var previous config.Profile
		if index >= 0 {
			previous = cfg.Profiles[index]
			if modelsOmitted {
				input.Models = append([]ModelInput(nil), configModelsToInput(previous.Models)...)
			}
			if modelsOmitted && input.DefaultModel == "" && previous.DefaultModel != "" {
				input.DefaultModel = previous.DefaultModel
			}
		}
		if index >= 0 && previous.Source == config.SourceDoge {
			if input.Source != config.SourceDoge {
				return errors.New("二狗子代理 API 来源不能修改")
			}
			// 二狗子密钥由远端令牌接口维护；后端再次校验，避免绕过前端只读控件修改。
			if normalizeDogeAPIKey(input.APIKey) != normalizeDogeAPIKey(previous.APIKey) {
				return errors.New("二狗子 API 密钥由远端管理，不能修改")
			}
			input.APIKey = normalizeDogeAPIKey(previous.APIKey)
		}
		profile := config.Profile{
			ID: input.ID, Source: input.Source, Category: input.Category, Name: input.Name, BaseURL: input.BaseURL,
			APIKey: input.APIKey, Note: input.Note, Headers: input.Headers, Models: modelInputsToConfig(input.Models), DefaultModel: input.DefaultModel,
		}
		if index >= 0 {
			profile.RemoteTokenID = previous.RemoteTokenID
		}
		if index < 0 {
			profile.ID = config.NewProfileID()
		}
		if profile.APIKey == "" {
			return errors.New("API 密钥不能为空")
		}
		if len([]rune(profile.Note)) > 160 {
			return errors.New("备注说明不能超过 160 个字符")
		}
		if err := config.ValidateProfile(profile); err != nil {
			return err
		}
		if index >= 0 {
			cfg.Profiles[index] = profile
			if profile.Source == config.SourceDoge && profile.RemoteTokenID > 0 {
				for tokenIndex := range cfg.Doge.Tokens {
					if cfg.Doge.Tokens[tokenIndex].ID == profile.RemoteTokenID {
						cfg.Doge.Tokens[tokenIndex].Category = profile.Category
						break
					}
				}
			}
			if previous.Category != profile.Category && cfg.ActiveProfiles[previous.Category] == profile.ID {
				delete(cfg.ActiveProfiles, previous.Category)
			}
		} else {
			cfg.Profiles = append(cfg.Profiles, profile)
		}
		return nil
	})
}

func (s *DesktopService) DeleteProfile(id string) error {
	return s.updateConfig(func(cfg *config.AppConfig) error {
		index := config.FindProfileIndex(cfg.Profiles, id)
		if index < 0 {
			return errors.New("代理 API 不存在")
		}
		category := cfg.Profiles[index].Category
		cfg.Profiles = append(cfg.Profiles[:index], cfg.Profiles[index+1:]...)
		if cfg.ActiveProfiles[category] == id {
			delete(cfg.ActiveProfiles, category)
		}
		return nil
	})
}

func (s *DesktopService) ActivateProfile(id string) error {
	return s.updateConfig(func(cfg *config.AppConfig) error {
		index := config.FindProfileIndex(cfg.Profiles, id)
		if index >= 0 {
			profile := cfg.Profiles[index]
			if profile.Source == config.SourceDoge && profile.RemoteTokenID > 0 {
				for _, token := range cfg.Doge.Tokens {
					if token.ID == profile.RemoteTokenID && !dogeTokenAvailable(token, cfg.Doge.Groups) {
						return errors.New("二狗子令牌当前分组不可用，不能启用")
					}
				}
			}
			if cfg.ActiveProfiles == nil {
				cfg.ActiveProfiles = map[string]string{}
			}
			cfg.ActiveProfiles[profile.Category] = id
			return nil
		}
		return errors.New("代理 API 不存在")
	})
}

// DismissDogeTokenSwitch 在当前失败状态持续期间抑制令牌切换提示，并在失败恢复后的五分钟内继续抑制。
// 抑制状态仅保存在内存中，不修改便携配置；失败状态恢复后会自动清理。
func (s *DesktopService) DismissDogeTokenSwitch(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("令牌切换提示已失效")
	}
	s.switchMu.Lock()
	promptState, ok := s.switchPrompts[key]
	if ok {
		promptState.dismissed = true
		promptState.suppressedUntil = time.Now().Add(5 * time.Minute)
	}
	s.switchMu.Unlock()
	if !ok {
		return errors.New("令牌切换提示已失效")
	}
	s.notifyStateChanged()
	return nil
}

// SwitchDogeToken 在服务端重新校验提示、类别和可用状态后启用候选令牌。
// 前端只提交运行时提示 key 与远端令牌 ID，不能借此切换到其他类别或不可用令牌。
func (s *DesktopService) SwitchDogeToken(key string, tokenID int64) error {
	key = strings.TrimSpace(key)
	if key == "" || tokenID <= 0 {
		return errors.New("令牌切换参数无效")
	}
	prompt := s.buildDogeTokenSwitchPrompt()
	if prompt == nil || prompt.Key != key {
		return errors.New("令牌切换提示已失效，请重新等待下一次异常")
	}
	state := s.runtime.State()
	if state == nil {
		return errors.New("程序尚未初始化")
	}
	profileID := state.Config.ActiveProfiles[prompt.Category]
	profileIndex := config.FindProfileIndex(state.Config.Profiles, profileID)
	if profileIndex < 0 || state.Config.Profiles[profileIndex].Source != config.SourceDoge {
		return errors.New("当前类别没有活动的二狗子 API")
	}
	currentProfile := state.Config.Profiles[profileIndex]
	var currentToken config.DogeToken
	for _, token := range state.Config.Doge.Tokens {
		if token.ID == prompt.CurrentTokenID {
			currentToken = token
			break
		}
	}
	directoryContext := s.dogeDirectorySwitchContext()
	directoryPrompt := prompt.FailureKind == "directory" && directoryContext != nil && directoryContext.key == key
	if directoryPrompt && currentToken.ID == 0 {
		currentToken = directoryContext.token
	}
	if currentToken.ID == 0 || currentToken.ID != currentProfile.RemoteTokenID {
		return errors.New("当前令牌已刷新，请重新等待下一次异常")
	}
	var candidate config.DogeToken
	found := false
	for _, token := range state.Config.Doge.Tokens {
		if token.ID == tokenID {
			candidate, found = token, true
			break
		}
	}
	if !found || candidate.ID == prompt.CurrentTokenID {
		return errors.New("候选令牌不存在")
	}
	if candidate.Category == "" {
		for _, imported := range state.Config.Profiles {
			if imported.Source == config.SourceDoge && imported.RemoteTokenID == candidate.ID {
				candidate.Category = imported.Category
				break
			}
		}
	}
	profilesByRemoteID := make(map[int64]config.Profile)
	for _, profile := range state.Config.Profiles {
		if profile.Source == config.SourceDoge && profile.RemoteTokenID > 0 {
			profilesByRemoteID[profile.RemoteTokenID] = profile
		}
	}
	validCandidates := availableDogeTokensForCategory(state.Config.Doge.Tokens, profilesByRemoteID, state.Config.Doge.Groups, currentProfile.Category)
	valid := false
	for _, validCandidate := range validCandidates {
		if validCandidate.ID == candidate.ID {
			candidate = validCandidate
			valid = true
			break
		}
	}
	if !valid {
		return errors.New("候选令牌当前不可用")
	}
	if err := s.EnableDogeToken(candidate.ID); err != nil {
		return err
	}
	s.runtime.ResetProfileHealth(currentProfile.ID)
	s.switchMu.Lock()
	delete(s.switchPrompts, key)
	if s.directorySwitch != nil && s.directorySwitch.key == key {
		s.directorySwitch = nil
	}
	s.switchMu.Unlock()
	s.notifyStateChanged()
	return nil
}

func (s *DesktopService) SetNetwork(input network.Settings) error {
	input.Mode = strings.TrimSpace(input.Mode)
	input.ProxyURL = strings.TrimSpace(input.ProxyURL)
	return s.updateConfig(func(cfg *config.AppConfig) error {
		if err := network.Validate(input, cfg.ProxyPort); err != nil {
			return err
		}
		cfg.Network = input
		return nil
	})
}

// SetProxyPort 校验并热切换本地代理监听端口；新端口无法绑定时不修改配置和现有监听。
// 端口范围为 TCP 的 1-65535；成功后新请求地址立即使用新端口，已有连接由旧服务优雅退出。
func (s *DesktopService) SetProxyPort(port int) error {
	if port < 1 || port > 65535 {
		return errors.New("监听端口必须是 1 到 65535 之间的整数")
	}
	state := s.runtime.State()
	if state == nil {
		return errors.New("程序尚未初始化")
	}
	if state.Config.ProxyPort == port {
		return nil
	}
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("代理端口 %s 无法监听，可能已被其他程序占用: %w", address, err)
	}
	server := &http.Server{
		Handler: s.runtime.ProxyHandler(), ReadHeaderTimeout: 15 * time.Second,
		IdleTimeout: 120 * time.Second,
	}
	if err := s.updateConfig(func(cfg *config.AppConfig) error {
		cfg.ProxyPort = port
		return nil
	}); err != nil {
		_ = listener.Close()
		return err
	}
	s.mu.Lock()
	oldServer := s.server
	oldListener := s.listener
	if oldServer == nil && oldListener == nil {
		s.mu.Unlock()
		_ = listener.Close()
		return nil
	}
	s.server = server
	s.listener = listener
	s.mu.Unlock()
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			application.Get().Logger.Error("透明代理异常退出", "error", err)
		}
	}()
	if oldServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := oldServer.Shutdown(ctx); err != nil {
			application.Get().Logger.Error("旧监听端口关闭失败", "error", err)
		}
		cancel()
	} else if oldListener != nil {
		_ = oldListener.Close()
	}
	return nil
}

func (s *DesktopService) SetPreferences(input config.Preferences) error {
	state := s.runtime.State()
	if state == nil {
		return errors.New("程序尚未初始化")
	}
	previous := state.Config.Preferences.LaunchAtStartup
	if err := platform.SetLaunchAtStartup(input.LaunchAtStartup); err != nil {
		return fmt.Errorf("更新开机启动失败: %w", err)
	}
	_, err := s.runtime.UpdateConfig(func(cfg *config.AppConfig) error {
		cfg.Preferences = input
		return nil
	})
	if err != nil {
		_ = platform.SetLaunchAtStartup(previous)
		return err
	}
	s.notifyStateChanged()
	return nil
}

// OpenDogeTopup 使用系统默认浏览器打开二狗子购买入口。
// 购买地址来自最近一次 `/api/user/topup/info` 同步结果，不接受前端传入的任意 URL。
func (s *DesktopService) OpenDogeTopup() error {
	state := s.runtime.State()
	if state == nil {
		return errors.New("程序尚未初始化")
	}
	topupLink := strings.TrimSpace(state.Config.Doge.Topup.TopupLink)
	if topupLink == "" {
		return errors.New("二狗子暂未提供购买入口")
	}
	return platform.OpenURL(topupLink)
}

// OpenDogeProfile 使用系统默认浏览器打开二狗子用户中心，令牌生成路径由界面说明固定提示。
func (s *DesktopService) OpenDogeProfile() error {
	return platform.OpenURL(dogeProfileURL)
}

func (s *DesktopService) TestProfile(id string) (TestResult, error) {
	state := s.runtime.State()
	index := config.FindProfileIndex(state.Config.Profiles, id)
	if index < 0 {
		return TestResult{}, errors.New("代理 API 不存在")
	}
	profile := config.CloneProfile(state.Config.Profiles[index])
	target, _ := url.Parse(profile.BaseURL)
	target.Path = relay.JoinTargetPath(target.Path, "/v1/models")
	transport, err := network.BuildTransport(state.Config.Network, network.DetectSystemProxy(), state.Config.ProxyPort)
	if err != nil {
		return TestResult{}, err
	}
	client := &http.Client{Transport: transport, Timeout: 15 * time.Second}
	request, _ := http.NewRequest(http.MethodGet, target.String(), nil)
	request.Header.Set("Authorization", "Bearer "+profile.APIKey)
	for name, value := range profile.Headers {
		request.Header.Set(name, value)
	}
	started := time.Now()
	response, err := client.Do(request)
	if err != nil {
		return TestResult{}, fmt.Errorf("连接失败: %s", relay.SanitizeError(err))
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 32*1024))
	return TestResult{
		OK: response.StatusCode >= 200 && response.StatusCode < 300, Reachable: true,
		Status: response.StatusCode, DurationMs: time.Since(started).Milliseconds(), URL: target.String(),
	}, nil
}

func publicProfile(profile config.Profile, activeProfiles map[string]string) PublicProfile {
	u, _ := url.Parse(profile.BaseURL)
	preview := ""
	if u != nil {
		u.Path = relay.JoinTargetPath(u.Path, "/v1/responses")
		preview = u.String()
	}
	return PublicProfile{
		ID: profile.ID, Source: profile.Source, Category: profile.Category, Name: profile.Name, BaseURL: profile.BaseURL,
		APIKey: profile.APIKey, Note: profile.Note, Headers: profile.Headers, Models: publicModels(profile.Models), DefaultModel: profile.DefaultModel,
		Active: activeProfiles[profile.Category] == profile.ID, PreviewURL: preview, RemoteTokenID: profile.RemoteTokenID,
	}
}

func maskDogeKey(key string) string {
	key = strings.TrimSpace(key)
	if len(key) <= 8 {
		return key
	}
	return key[:4] + "**********" + key[len(key)-4:]
}

// dogeTokenNote 为首次同步的令牌生成可读备注；只使用接口返回的掩码密钥和额度摘要，不写入完整密钥。
func dogeTokenNote(token config.DogeToken) string {
	key := strings.TrimSpace(token.MaskedKey)
	if key == "" {
		key = maskDogeKey(token.Key)
	}
	if key == "" {
		return ""
	}
	quota := "不限额度"
	if !token.UnlimitedQuota {
		quota = "剩余 " + strconv.FormatInt(token.RemainQuota, 10)
	}
	return normalizeDogeAPIKey(key) + " · " + quota
}
