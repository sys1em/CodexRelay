/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : 上游代理失败次数的运行时观察
 * @File          : 令牌错误 streak 与短窗口异常统计
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package relay

import "time"

const (
	AuthFailureThreshold     = 5
	UpstreamFailureThreshold = 5
	UpstreamFailureWindow    = 3 * time.Minute
)

// HealthSnapshot 是单个本地代理 Profile 的运行时错误摘要，不写入 usage.json 或 config.json。
type HealthSnapshot struct {
	ProfileID         string    `json:"profileId"`
	Category          string    `json:"category"`
	AuthFailures      int       `json:"authFailures"`
	UpstreamFailures  int       `json:"upstreamFailures"`
	AuthTriggered     bool      `json:"authTriggered"`
	UpstreamTriggered bool      `json:"upstreamTriggered"`
	TriggerGeneration uint64    `json:"triggerGeneration"`
	LastStatus        int       `json:"lastStatus"`
	LastFailureAt     time.Time `json:"lastFailureAt"`
}

type profileHealthState struct {
	snapshot           HealthSnapshot
	upstreamErrors     []time.Time
	authGeneration     uint64
	upstreamGeneration uint64
}

// SetHealthChangedHandler 注册错误阈值变化通知；回调不在健康状态锁内执行。
func (r *Runtime) SetHealthChangedHandler(handler func()) {
	r.healthMu.Lock()
	r.healthChanged = handler
	r.healthMu.Unlock()
}

// ObserveUpstreamResult 记录一次已收到的上游响应或连接错误。
// 401/403 只统计连续失败；5xx 和传输错误在固定时间窗内累计，其他状态不触发令牌切换提醒。
func (r *Runtime) ObserveUpstreamResult(profileID, category string, status int, transportError bool) HealthSnapshot {
	if profileID == "" {
		return HealthSnapshot{}
	}
	now := time.Now()
	r.healthMu.Lock()
	if r.health == nil {
		r.health = make(map[string]*profileHealthState)
	}
	entry := r.health[profileID]
	if entry == nil {
		entry = &profileHealthState{snapshot: HealthSnapshot{ProfileID: profileID, Category: category}}
		r.health[profileID] = entry
	}
	previousAuth := entry.snapshot.AuthTriggered
	previousUpstream := entry.snapshot.UpstreamTriggered
	entry.snapshot.Category = category
	entry.snapshot.LastStatus = status
	entry.snapshot.LastFailureAt = now
	if status == 401 || status == 403 {
		entry.snapshot.AuthFailures++
	} else {
		entry.snapshot.AuthFailures = 0
	}
	if transportError || status >= 500 {
		entry.upstreamErrors = append(entry.upstreamErrors, now)
	}
	cutoff := now.Add(-UpstreamFailureWindow)
	kept := entry.upstreamErrors[:0]
	for _, at := range entry.upstreamErrors {
		if !at.Before(cutoff) {
			kept = append(kept, at)
		}
	}
	entry.upstreamErrors = kept
	entry.snapshot.UpstreamFailures = len(kept)
	entry.snapshot.AuthTriggered = entry.snapshot.AuthFailures >= AuthFailureThreshold
	entry.snapshot.UpstreamTriggered = entry.snapshot.UpstreamFailures >= UpstreamFailureThreshold
	if entry.snapshot.AuthTriggered && !previousAuth {
		entry.authGeneration++
	}
	if entry.snapshot.UpstreamTriggered && !previousUpstream {
		entry.upstreamGeneration++
	}
	if entry.snapshot.AuthTriggered {
		entry.snapshot.TriggerGeneration = entry.authGeneration
	} else if entry.snapshot.UpstreamTriggered {
		entry.snapshot.TriggerGeneration = entry.upstreamGeneration
	} else {
		entry.snapshot.TriggerGeneration = 0
	}
	changed := previousAuth != entry.snapshot.AuthTriggered || previousUpstream != entry.snapshot.UpstreamTriggered
	snapshot := entry.snapshot
	handler := r.healthChanged
	r.healthMu.Unlock()
	if changed && handler != nil {
		handler()
	}
	return snapshot
}

// HealthSnapshots 返回当前运行时错误摘要，并清理已超过统计窗口的 5xx/连接错误。
func (r *Runtime) HealthSnapshots() []HealthSnapshot {
	now := time.Now()
	r.healthMu.Lock()
	defer r.healthMu.Unlock()
	result := make([]HealthSnapshot, 0, len(r.health))
	cutoff := now.Add(-UpstreamFailureWindow)
	for profileID, entry := range r.health {
		kept := entry.upstreamErrors[:0]
		for _, at := range entry.upstreamErrors {
			if !at.Before(cutoff) {
				kept = append(kept, at)
			}
		}
		entry.upstreamErrors = kept
		entry.snapshot.UpstreamFailures = len(kept)
		entry.snapshot.UpstreamTriggered = len(kept) >= UpstreamFailureThreshold
		if entry.snapshot.AuthTriggered {
			entry.snapshot.TriggerGeneration = entry.authGeneration
		} else if entry.snapshot.UpstreamTriggered {
			entry.snapshot.TriggerGeneration = entry.upstreamGeneration
		} else {
			entry.snapshot.TriggerGeneration = 0
		}
		if entry.snapshot.AuthFailures == 0 && len(kept) == 0 {
			delete(r.health, profileID)
			continue
		}
		result = append(result, entry.snapshot)
	}
	return result
}

// ResetProfileHealth 清除指定 Profile 的失败统计；切换成功或手动切换后调用。
func (r *Runtime) ResetProfileHealth(profileID string) {
	if profileID == "" {
		return
	}
	r.healthMu.Lock()
	_, existed := r.health[profileID]
	delete(r.health, profileID)
	handler := r.healthChanged
	r.healthMu.Unlock()
	if existed && handler != nil {
		handler()
	}
}
