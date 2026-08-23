/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : Token 观察与持久化回归测试
 */
package usage

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestObserverReadsChunkedResponsesSSE(t *testing.T) {
	observer := NewObserver(http.Header{"Content-Type": []string{"text/event-stream; charset=utf-8"}})
	chunks := [][]byte{
		[]byte("event: response.output_text.delta\ndata: {\"type\":\"response.output_text.delta\",\"delta\":\"hello\"}\n\n"),
		[]byte("event: response.completed\nda"),
		[]byte("ta: {\"type\":\"response.completed\",\"response\":{\"model\":\"gpt-5\",\"usage\":{\"input_tokens\":120,\"output_tokens\":30,"),
		[]byte("\"input_tokens_details\":{\"cached_tokens\":40},\"total_tokens\":150}}}\n\n"),
	}
	for _, chunk := range chunks {
		observer.Observe(chunk)
	}
	observed := observer.Finish()
	if observed.Status != StatusReported || observed.Model != "gpt-5" || observed.TotalTokens != 150 {
		t.Fatalf("usage = %+v", observed)
	}
}

func TestObserverReadsNonStreamingJSON(t *testing.T) {
	observer := NewObserver(http.Header{"Content-Type": []string{"application/json"}})
	observer.Observe([]byte(`{"model":"gpt-5-mini","usage":{"input_tokens":12,"output_tokens":8,"input_tokens_details":{"cached_tokens":2}}}`))
	observed := observer.Finish()
	if observed.Status != StatusReported || observed.TotalTokens != 20 || observed.CachedTokens != 2 {
		t.Fatalf("usage = %+v", observed)
	}
}

func TestObserverDoesNotInspectEncodedResponse(t *testing.T) {
	observer := NewObserver(http.Header{
		"Content-Type": []string{"text/event-stream"}, "Content-Encoding": []string{"gzip"},
	})
	observer.Observe([]byte(`data: {"type":"response.completed","response":{"usage":{"input_tokens":1,"output_tokens":2}}}\n\n`))
	if observed := observer.Finish(); observed.Status != StatusUnreported {
		t.Fatalf("usage = %+v, want unreported", observed)
	}
}

func TestStoreKeepsLast300RequestsAndAggregatesAll(t *testing.T) {
	store, err := NewStore(filepath.Join(t.TempDir(), "usage.json"))
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 305; index++ {
		record := RequestRecord{
			ID: uint64(index + 1), StartedAt: time.Now(), ProfileID: "profile-a",
			UsageStatus: StatusReported, InputTokens: 10, OutputTokens: 5, CachedTokens: 2, TotalTokens: 15,
		}
		if err := store.Record(record); err != nil {
			t.Fatal(err)
		}
	}
	snapshot := store.Snapshot()
	if len(snapshot.Recent) != 300 || snapshot.Recent[0].ID != 305 || snapshot.Recent[299].ID != 6 {
		t.Fatalf("unexpected recent records: first=%d last=%d count=%d", snapshot.Recent[0].ID, snapshot.Recent[299].ID, len(snapshot.Recent))
	}
	if snapshot.Total.Requests != 305 || snapshot.Total.TotalTokens != 4575 {
		t.Fatalf("aggregate = %+v", snapshot.Total)
	}
}

func TestStorePersistsPortableJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.json")
	store, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Record(RequestRecord{ID: 1, StartedAt: time.Now(), ProfileID: "one", UsageStatus: StatusUnreported}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Data
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Total.Requests != 1 || decoded.Total.ReportedRequests != 0 {
		t.Fatalf("usage total = %+v", decoded.Total)
	}
}

func TestDeferredStoreDoesNotCreateUntilActivated(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.json")
	store, err := NewDeferredStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("deferred load created usage file: %v", err)
	}
	if err := store.Record(RequestRecord{ID: 1, StartedAt: time.Now(), ProfileID: "one", UsageStatus: StatusUnreported}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("deferred record created usage file: %v", err)
	}
	if err := store.ActivatePersistence(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("activated store did not create usage file: %v", err)
	}
}
