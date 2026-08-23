/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : 上游失败阈值观察回归测试
 * @File          : 令牌认证失败与上游异常窗口测试
 */
package relay

import (
	"testing"
	"time"
)

func TestObserveUpstreamResultAuthThresholdAndReset(t *testing.T) {
	runtime := &Runtime{health: make(map[string]*profileHealthState)}
	for index := 0; index < AuthFailureThreshold-1; index++ {
		snapshot := runtime.ObserveUpstreamResult("profile", "codex", 403, false)
		if snapshot.AuthTriggered {
			t.Fatalf("auth prompt triggered after %d failures", index+1)
		}
	}
	snapshot := runtime.ObserveUpstreamResult("profile", "codex", 403, false)
	if !snapshot.AuthTriggered || snapshot.AuthFailures != AuthFailureThreshold {
		t.Fatalf("auth threshold snapshot = %+v", snapshot)
	}
	generation := snapshot.TriggerGeneration
	snapshot = runtime.ObserveUpstreamResult("profile", "codex", 200, false)
	if snapshot.AuthTriggered || snapshot.AuthFailures != 0 {
		t.Fatalf("successful request did not reset auth streak = %+v", snapshot)
	}
	for index := 0; index < AuthFailureThreshold; index++ {
		snapshot = runtime.ObserveUpstreamResult("profile", "codex", 401, false)
	}
	if !snapshot.AuthTriggered || snapshot.TriggerGeneration <= generation {
		t.Fatalf("new auth failure should start a new generation = %+v", snapshot)
	}
}

func TestObserveUpstreamResultWindowThreshold(t *testing.T) {
	runtime := &Runtime{health: make(map[string]*profileHealthState)}
	for index := 0; index < UpstreamFailureThreshold-1; index++ {
		snapshot := runtime.ObserveUpstreamResult("profile", "codex", 502, false)
		if snapshot.UpstreamTriggered {
			t.Fatalf("upstream prompt triggered after %d failures", index+1)
		}
	}
	snapshot := runtime.ObserveUpstreamResult("profile", "codex", 502, false)
	if !snapshot.UpstreamTriggered || snapshot.UpstreamFailures != UpstreamFailureThreshold {
		t.Fatalf("upstream threshold snapshot = %+v", snapshot)
	}
	runtime.healthMu.Lock()
	entry := runtime.health["profile"]
	entry.upstreamErrors[0] = time.Now().Add(-UpstreamFailureWindow - time.Second)
	runtime.healthMu.Unlock()
	snapshots := runtime.HealthSnapshots()
	if len(snapshots) != 1 || snapshots[0].UpstreamTriggered || snapshots[0].UpstreamFailures != UpstreamFailureThreshold-1 {
		t.Fatalf("expired upstream failure was not pruned = %+v", snapshots)
	}
}
