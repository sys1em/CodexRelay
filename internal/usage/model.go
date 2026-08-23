/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : Token 请求明细与聚合模型
 */
package usage

import "time"

const (
	StatusReported   = "reported"
	StatusUnreported = "unreported"
)

type RequestRecord struct {
	ID           uint64    `json:"id"`
	StartedAt    time.Time `json:"startedAt"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	ProfileID    string    `json:"profileId,omitempty"`
	Profile      string    `json:"profile,omitempty"`
	Status       int       `json:"status"`
	Duration     int64     `json:"durationMs"`
	Error        string    `json:"error,omitempty"`
	UsageStatus  string    `json:"usageStatus"`
	Model        string    `json:"model,omitempty"`
	InputTokens  uint64    `json:"inputTokens,omitempty"`
	OutputTokens uint64    `json:"outputTokens,omitempty"`
	CachedTokens uint64    `json:"cachedTokens,omitempty"`
	TotalTokens  uint64    `json:"totalTokens,omitempty"`
}

type Aggregate struct {
	Requests         uint64 `json:"requests"`
	ReportedRequests uint64 `json:"reportedRequests"`
	InputTokens      uint64 `json:"inputTokens"`
	OutputTokens     uint64 `json:"outputTokens"`
	CachedTokens     uint64 `json:"cachedTokens"`
	TotalTokens      uint64 `json:"totalTokens"`
}

type Day struct {
	Total    Aggregate            `json:"total"`
	Profiles map[string]Aggregate `json:"profiles"`
}

type Data struct {
	Version   int                  `json:"version"`
	UpdatedAt time.Time            `json:"updatedAt"`
	Total     Aggregate            `json:"total"`
	Profiles  map[string]Aggregate `json:"profiles"`
	Days      map[string]Day       `json:"days"`
	Recent    []RequestRecord      `json:"recent"`
}

type Overview struct {
	UpdatedAt time.Time            `json:"updatedAt"`
	Total     Aggregate            `json:"total"`
	Profiles  map[string]Aggregate `json:"profiles"`
	Days      map[string]Day       `json:"days"`
}
