/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : 透明代理与流式响应回归测试
 */
package relay

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codexrelay/internal/config"
	"codexrelay/internal/usage"
)

func TestJoinTargetPath(t *testing.T) {
	tests := []struct{ base, request, want string }{
		{"", "/v1/responses", "/v1/responses"},
		{"/v1", "/v1/responses", "/v1/responses"},
		{"/v1", "/responses", "/v1/responses"},
		{"/api/openai/v1", "/v1/responses", "/api/openai/v1/responses"},
		{"/v1/responses", "/v1/responses", "/v1/responses"},
		{"/v1/responses", "/responses/compact", "/v1/responses/compact"},
	}
	for _, test := range tests {
		if got := JoinTargetPath(test.base, test.request); got != test.want {
			t.Errorf("JoinTargetPath(%q, %q) = %q, want %q", test.base, test.request, got, test.want)
		}
	}
}

func TestRoutePathUsesCategoryPrefixAndRejectsLegacyV1(t *testing.T) {
	category, remainder, ok := RoutePath("/codex/responses")
	if !ok || category != config.CategoryCodex || remainder != "/responses" {
		t.Fatalf("route = %q %q %v", category, remainder, ok)
	}
	if _, _, ok := RoutePath("/v1/responses"); ok {
		t.Fatal("legacy /v1 route must be rejected")
	}
	for _, prefix := range []string{"/image/generations", "/other/request"} {
		category, remainder, ok := RoutePath(prefix)
		if !ok || remainder == "" || (prefix == "/image/generations" && category != config.CategoryImage) || (prefix == "/other/request" && category != config.CategoryOther) {
			t.Fatalf("route %q = %q %q %v", prefix, category, remainder, ok)
		}
	}
}

func TestProxyPreservesBodyAndReplacesAuthorization(t *testing.T) {
	body := []byte{0x28, 0xb5, 0x2f, 0xfd, 0x00, 0x11, 0x22, 0xff}
	type capturedRequest struct {
		method, path, query, authorization, contentType string
		body                                            []byte
	}
	captured := make(chan capturedRequest, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		data, _ := io.ReadAll(request.Body)
		captured <- capturedRequest{request.Method, request.URL.Path, request.URL.RawQuery, request.Header.Get("Authorization"), request.Header.Get("Content-Type"), data}
		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("X-Upstream-Test", "preserved")
		writer.WriteHeader(http.StatusCreated)
		_, _ = writer.Write([]byte(`{"ok":true}`))
	}))
	defer upstream.Close()

	runtime := testRuntime(t, upstream.URL+"/v1", "upstream-secret")
	proxy := httptest.NewServer(runtime.ProxyHandler())
	defer proxy.Close()
	req, _ := http.NewRequest(http.MethodPost, proxy.URL+"/codex/responses?trace=1", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer sk-local-token")
	req.Header.Set("Content-Type", "application/octet-stream")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	responseBody, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	got := <-captured
	if response.StatusCode != http.StatusCreated || got.method != http.MethodPost || got.path != "/v1/responses" || got.query != "trace=1" {
		t.Fatalf("unexpected request/status: %+v %d", got, response.StatusCode)
	}
	if response.Header.Get("X-Upstream-Test") != "preserved" || !bytes.Equal(responseBody, []byte(`{"ok":true}`)) {
		t.Fatalf("forwarded response changed: header=%q body=%q", response.Header.Get("X-Upstream-Test"), responseBody)
	}
	if got.authorization != "Bearer upstream-secret" || got.contentType != "application/octet-stream" || !bytes.Equal(got.body, body) {
		t.Fatalf("forwarded request changed: %+v", got)
	}
}

func TestProxyStreamsSSEBeforeUpstreamCompletes(t *testing.T) {
	release := make(chan struct{})
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(writer, "data: first\n\n")
		writer.(http.Flusher).Flush()
		<-release
		_, _ = io.WriteString(writer, "data: second\n\n")
	}))
	defer upstream.Close()
	runtime := testRuntime(t, upstream.URL, "secret")
	proxy := httptest.NewServer(runtime.ProxyHandler())
	defer proxy.Close()
	req, _ := http.NewRequest(http.MethodPost, proxy.URL+"/codex/responses", strings.NewReader(`{"stream":true}`))
	req.Header.Set("Authorization", "Bearer sk-local-token")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	lineCh := make(chan string, 1)
	go func() { line, _ := bufio.NewReader(response.Body).ReadString('\n'); lineCh <- line }()
	select {
	case line := <-lineCh:
		if line != "data: first\n" {
			t.Fatalf("first line = %q", line)
		}
	case <-time.After(time.Second):
		t.Fatal("first SSE event was buffered")
	}
	close(release)
}

func TestProxyPreservesSSEBytesAndRecordsUsage(t *testing.T) {
	expected := []byte("event: response.created\ndata: {\"type\":\"response.created\"}\n\n" +
		"event: response.completed\ndata: {\"type\":\"response.completed\",\"response\":{\"model\":\"gpt-5.4\",\"usage\":{\"input_tokens\":101,\"output_tokens\":22,\"input_tokens_details\":{\"cached_tokens\":64},\"total_tokens\":123}}}\n\n")
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/event-stream")
		for _, chunk := range [][]byte{expected[:37], expected[37:109], expected[109:]} {
			_, _ = writer.Write(chunk)
			writer.(http.Flusher).Flush()
		}
	}))
	defer upstream.Close()
	runtime := testRuntime(t, upstream.URL, "secret")
	proxy := httptest.NewServer(runtime.ProxyHandler())
	defer proxy.Close()
	req, _ := http.NewRequest(http.MethodPost, proxy.URL+"/codex/responses", strings.NewReader(`{"stream":true}`))
	req.Header.Set("Authorization", "Bearer sk-local-token")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil || !bytes.Equal(actual, expected) {
		t.Fatalf("SSE response changed: %v\nactual=%q\nwant=%q", err, actual, expected)
	}
	records := runtime.RecentRecords()
	if len(records) != 1 || records[0].UsageStatus != usage.StatusReported || records[0].TotalTokens != 123 {
		t.Fatalf("records = %+v", records)
	}
}

func TestImageCategoryRecordsRequestWithoutTokenUsage(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(writer, `{"model":"image-model","usage":{"input_tokens":101,"output_tokens":22,"total_tokens":123}}`)
	}))
	defer upstream.Close()
	runtime := testRuntimeForCategory(t, upstream.URL, "secret", config.CategoryImage)
	proxy := httptest.NewServer(runtime.ProxyHandler())
	defer proxy.Close()
	req, _ := http.NewRequest(http.MethodPost, proxy.URL+"/image/generations", strings.NewReader(`{"prompt":"test"}`))
	req.Header.Set("Authorization", "Bearer sk-local-token")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, response.Body)
	response.Body.Close()
	records := runtime.RecentRecords()
	if len(records) != 1 || records[0].UsageStatus != usage.StatusUnreported || records[0].TotalTokens != 0 || records[0].InputTokens != 0 || records[0].OutputTokens != 0 {
		t.Fatalf("records = %+v", records)
	}
	if records[0].Status != http.StatusOK {
		t.Fatalf("status = %d", records[0].Status)
	}
}

func TestWrongLocalTokenIsRejected(t *testing.T) {
	runtime := testRuntime(t, "http://127.0.0.1:1", "secret")
	req := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer wrong-token")
	recorder := httptest.NewRecorder()
	runtime.ProxyHandler().ServeHTTP(recorder, req)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestUnconfiguredCategoryReturnsUnavailable(t *testing.T) {
	runtime := testRuntime(t, "http://127.0.0.1:1", "secret")
	req := httptest.NewRequest(http.MethodPost, "/claude/messages", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer sk-local-token")
	recorder := httptest.NewRecorder()
	runtime.ProxyHandler().ServeHTTP(recorder, req)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func testRuntime(t *testing.T, baseURL, apiKey string) *Runtime {
	return testRuntimeForCategory(t, baseURL, apiKey, config.CategoryCodex)
}

func testRuntimeForCategory(t *testing.T, baseURL, apiKey, category string) *Runtime {
	t.Helper()
	directory := t.TempDir()
	cfg := config.Default(18765)
	cfg.LocalAccessToken = "sk-local-token"
	cfg.Network.Mode = "direct"
	cfg.ActiveProfiles[category] = "test"
	cfg.Profiles = []config.Profile{{ID: "test", Source: config.SourceCustom, Category: category, Name: "Test", BaseURL: baseURL, APIKey: apiKey}}
	configStore := config.NewStore(filepath.Join(directory, "config.json"))
	usageStore, err := usage.NewStore(filepath.Join(directory, "usage.json"))
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := New(configStore, usageStore, cfg)
	if err != nil {
		t.Fatal(err)
	}
	return runtime
}
