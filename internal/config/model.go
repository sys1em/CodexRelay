/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : 配置模型、克隆与标识生成
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"codexrelay/internal/network"
)

const DefaultProxyPort = 8765

// DefaultDogeSyncIntervalMinutes 是新配置的二狗子自动同步默认间隔。
const DefaultDogeSyncIntervalMinutes = 3

const (
	SourceDoge   = "doge"
	SourceCustom = "custom"

	// RestoreViewCurrent 恢复窗口时保留用户当前的主页来源和类别筛选。
	RestoreViewCurrent = "current"
	// RestoreViewDefault 恢复窗口时重新应用 Preferences 中的默认来源和类别。
	RestoreViewDefault = "default"

	CategoryCodex    = "codex"
	CategoryClaude   = "claude"
	CategoryGemini   = "gemini"
	CategoryGrok     = "grok"
	CategoryOpenCode = "opencode"
	CategoryOpenClaw = "openclaw"
	CategoryHermes   = "hermes"
	CategoryImage    = "image"
	CategoryOther    = "other"
)

var Categories = []string{
	CategoryCodex,
	CategoryClaude,
	CategoryGemini,
	CategoryGrok,
	CategoryOpenCode,
	CategoryOpenClaw,
	CategoryHermes,
	CategoryImage,
	CategoryOther,
}

// ClientConfig 保存外部 AI 客户端的配置目录和主配置文件名。
// 路径由启动时按已知默认目录只读探测，用户也可以在高级设置中覆盖；不保存外部文件正文。
type ClientConfig struct {
	ConfigDir             string `json:"configDir,omitempty"`
	ConfigFile            string `json:"configFile,omitempty"`
	SkipConfigReplacement bool   `json:"skipConfigReplacement,omitempty"`
}

// ModelEntry 保存一个代理 API 可用的上游模型；目录来自用户主动获取或手动维护，不推断模型能力。
type ModelEntry struct {
	ID            string `json:"id"`
	Name          string `json:"name,omitempty"`
	OwnedBy       string `json:"ownedBy,omitempty"`
	ContextWindow int64  `json:"contextWindow,omitempty"`
}

type Profile struct {
	ID            string            `json:"id"`
	Source        string            `json:"source"`
	Category      string            `json:"category"`
	Name          string            `json:"name"`
	BaseURL       string            `json:"baseUrl"`
	APIKey        string            `json:"apiKey"`
	Note          string            `json:"note,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	Models        []ModelEntry      `json:"models,omitempty"`
	DefaultModel  string            `json:"defaultModel,omitempty"`
	RemoteTokenID int64             `json:"remoteTokenId,omitempty"`
}

// DogeToken 保存二狗子 New API 管理目录；Key 在绑定或手动同步时从密钥接口补全，自动同步只为新增且缺失的令牌补全。
type DogeToken struct {
	ID                 int64   `json:"id"`
	UserID             int64   `json:"userId,omitempty"`
	Key                string  `json:"key,omitempty"`
	MaskedKey          string  `json:"maskedKey,omitempty"`
	Status             int     `json:"status"`
	Name               string  `json:"name"`
	CreatedTime        int64   `json:"createdTime,omitempty"`
	AccessedTime       int64   `json:"accessedTime,omitempty"`
	ExpiredTime        int64   `json:"expiredTime,omitempty"`
	RemainQuota        int64   `json:"remainQuota,omitempty"`
	UnlimitedQuota     bool    `json:"unlimitedQuota"`
	ModelLimitsEnabled bool    `json:"modelLimitsEnabled"`
	ModelLimits        string  `json:"modelLimits,omitempty"`
	AllowIPs           string  `json:"allowIps,omitempty"`
	UsedQuota          int64   `json:"usedQuota,omitempty"`
	Group              string  `json:"group"`                      // 二狗子令牌接口返回的原始分组键，仅用于权限和同组判断。
	GroupDisplayName   string  `json:"groupDisplayName,omitempty"` // /api/user/self/groups 的 display_name，只用于用户可见文案。
	GroupRatio         float64 `json:"groupRatio,omitempty"`
	CrossGroupRetry    bool    `json:"crossGroupRetry"`
	Category           string  `json:"category,omitempty"`
	Note               string  `json:"note,omitempty"`
}

// DogeAccount 保存二狗子当前用户的余额与身份摘要；quota 为上游原始额度单位。
type DogeAccount struct {
	ID           int64  `json:"id,omitempty"`
	Username     string `json:"username,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	Email        string `json:"email,omitempty"`
	Group        string `json:"group,omitempty"`
	Status       int    `json:"status,omitempty"`
	Quota        int64  `json:"quota,omitempty"`
	UsedQuota    int64  `json:"usedQuota,omitempty"`
	RequestCount int64  `json:"requestCount,omitempty"`
}

// DogeSubscription 保存当前有效套餐的额度、状态和到期时间；金额字段为上游原始额度单位。
type DogeSubscription struct {
	ID          int64  `json:"id"`
	PlanID      int64  `json:"planId"`
	PlanTitle   string `json:"planTitle"`
	Status      string `json:"status"`
	AmountTotal int64  `json:"amountTotal"`
	AmountUsed  int64  `json:"amountUsed"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
}

// DogeTopupInfo 保存兑换开关与购买入口；不保存兑换记录中的兑换 key。
type DogeTopupInfo struct {
	EnableRedemption bool   `json:"enableRedemption"`
	TopupLink        string `json:"topupLink,omitempty"`
}

// DogeAnnouncement 保存公告接口返回的公开内容；ID 是公告在上游的稳定身份。
type DogeAnnouncement struct {
	ID          int64  `json:"id"`
	Content     string `json:"content"`
	Extra       string `json:"extra,omitempty"`
	PublishDate string `json:"publishDate"`
	Type        string `json:"type"`
}

// DogeNotificationState 保存公告缓存、未读状态和一次性提醒确认状态。
type DogeNotificationState struct {
	Initialized               bool               `json:"initialized"`
	AnnouncementsEnabled      bool               `json:"announcementsEnabled"`
	CurrentNotice             string             `json:"currentNotice,omitempty"`
	Announcements             []DogeAnnouncement `json:"announcements"`
	ReadAnnouncementIDs       []int64            `json:"readAnnouncementIds"`
	DismissedAlertKeys        []string           `json:"dismissedAlertKeys"`
	LastAnnouncementSyncAt    time.Time          `json:"lastAnnouncementSyncAt,omitempty"`
	LastAnnouncementSyncError string             `json:"lastAnnouncementSyncError,omitempty"`
}

type DogeConnection struct {
	BaseURL             string                `json:"baseUrl"`
	AccessToken         string                `json:"accessToken,omitempty"`
	SyncIntervalMinutes int                   `json:"syncIntervalMinutes"`
	User                map[string]any        `json:"user,omitempty"`
	Account             DogeAccount           `json:"account"`
	Subscriptions       []DogeSubscription    `json:"subscriptions"`
	Topup               DogeTopupInfo         `json:"topup"`
	Notifications       DogeNotificationState `json:"notifications"`
	Groups              []string              `json:"groups"`
	Tokens              []DogeToken           `json:"tokens"`
	TokenOrder          []string              `json:"tokenOrder"`
	LastSyncAt          time.Time             `json:"lastSyncAt,omitempty"`
	LastSyncError       string                `json:"lastSyncError,omitempty"`
}

// Preferences 保存窗口生命周期、主页类别可见性、默认筛选和恢复策略；不会改变代理运行时路由。
type Preferences struct {
	CloseToTray       bool     `json:"closeToTray"`
	LaunchAtStartup   bool     `json:"launchAtStartup"`
	StartHidden       bool     `json:"startHidden"`
	VisibleCategories []string `json:"visibleCategories"`
	DefaultSource     string   `json:"defaultSource"`
	DefaultCategory   string   `json:"defaultCategory"`
	RestoreViewMode   string   `json:"restoreViewMode"`
}

type AppConfig struct {
	ProxyPort        int                     `json:"proxyPort"`
	LocalAccessToken string                  `json:"localAccessToken"`
	ActiveProfiles   map[string]string       `json:"activeProfiles"`
	Profiles         []Profile               `json:"profiles"`
	ClientConfigs    map[string]ClientConfig `json:"clientConfigs"`
	Network          network.Settings        `json:"network"`
	Preferences      Preferences             `json:"preferences"`
	Doge             DogeConnection          `json:"doge"`
}

func Default(proxyPort int) AppConfig {
	if proxyPort == 0 {
		proxyPort = DefaultProxyPort
	}
	return AppConfig{
		ProxyPort:        proxyPort,
		LocalAccessToken: newLocalAccessToken(),
		Profiles:         []Profile{},
		ActiveProfiles:   map[string]string{},
		ClientConfigs:    map[string]ClientConfig{},
		Network:          network.Settings{Mode: "system"},
		Preferences:      Preferences{CloseToTray: true, VisibleCategories: append([]string(nil), Categories...), DefaultSource: SourceDoge, DefaultCategory: CategoryCodex, RestoreViewMode: RestoreViewCurrent},
		Doge:             DogeConnection{BaseURL: "https://api.ergouzi.life", SyncIntervalMinutes: DefaultDogeSyncIntervalMinutes, Notifications: DogeNotificationState{Announcements: []DogeAnnouncement{}, ReadAnnouncementIDs: []int64{}, DismissedAlertKeys: []string{}}, Groups: []string{}, Tokens: []DogeToken{}, TokenOrder: []string{}},
	}
}

func Clone(source AppConfig) AppConfig {
	clone := source
	clone.Preferences.VisibleCategories = append([]string(nil), source.Preferences.VisibleCategories...)
	clone.Profiles = make([]Profile, len(source.Profiles))
	for i := range source.Profiles {
		clone.Profiles[i] = CloneProfile(source.Profiles[i])
	}
	clone.ActiveProfiles = make(map[string]string, len(source.ActiveProfiles))
	for category, id := range source.ActiveProfiles {
		clone.ActiveProfiles[category] = id
	}
	clone.ClientConfigs = make(map[string]ClientConfig, len(source.ClientConfigs))
	for category, client := range source.ClientConfigs {
		clone.ClientConfigs[category] = client
	}
	clone.Doge.Groups = append([]string(nil), source.Doge.Groups...)
	clone.Doge.Tokens = make([]DogeToken, len(source.Doge.Tokens))
	copy(clone.Doge.Tokens, source.Doge.Tokens)
	clone.Doge.TokenOrder = append([]string(nil), source.Doge.TokenOrder...)
	clone.Doge.Subscriptions = append([]DogeSubscription(nil), source.Doge.Subscriptions...)
	clone.Doge.Notifications.Announcements = append([]DogeAnnouncement(nil), source.Doge.Notifications.Announcements...)
	clone.Doge.Notifications.ReadAnnouncementIDs = append([]int64(nil), source.Doge.Notifications.ReadAnnouncementIDs...)
	clone.Doge.Notifications.DismissedAlertKeys = append([]string(nil), source.Doge.Notifications.DismissedAlertKeys...)
	if source.Doge.User != nil {
		clone.Doge.User = make(map[string]any, len(source.Doge.User))
		for key, value := range source.Doge.User {
			clone.Doge.User[key] = value
		}
	}
	return clone
}

func CloneProfile(source Profile) Profile {
	clone := source
	if source.Headers != nil {
		clone.Headers = make(map[string]string, len(source.Headers))
		for key, value := range source.Headers {
			clone.Headers[key] = value
		}
	}
	clone.Models = append([]ModelEntry(nil), source.Models...)
	return clone
}

func FindProfileIndex(profiles []Profile, id string) int {
	if id == "" {
		return -1
	}
	for i := range profiles {
		if profiles[i].ID == id {
			return i
		}
	}
	return -1
}

func OrderProfiles(profiles []Profile, ids []string) ([]Profile, error) {
	if len(ids) != len(profiles) {
		return nil, errors.New("排序列表与中转站数量不一致")
	}
	byID := make(map[string]Profile, len(profiles))
	for _, profile := range profiles {
		byID[profile.ID] = profile
	}
	ordered := make([]Profile, 0, len(ids))
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			return nil, errors.New("排序列表包含重复中转站")
		}
		profile, ok := byID[id]
		if !ok {
			return nil, errors.New("排序列表包含未知中转站")
		}
		seen[id] = true
		ordered = append(ordered, profile)
	}
	return ordered, nil
}

func NewProfileID() string {
	return randomToken(12)
}

// IsDogeSyncInterval 判断二狗子自动同步间隔是否为界面和服务端共同支持的分钟数。
func IsDogeSyncInterval(minutes int) bool {
	switch minutes {
	case 1, 3, 5, 10, 15, 30, 60:
		return true
	default:
		return false
	}
}

func randomToken(bytes int) string {
	buffer := make([]byte, bytes)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("relay-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buffer)
}

func newLocalAccessToken() string {
	return "sk-" + randomToken(24)
}
