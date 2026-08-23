/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : 模型目录获取与响应解析回归测试
 * @File          : 模型目录服务测试
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package desktop

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"codexrelay/internal/config"
)

func TestModelEndpointCandidatesUseVersionAndRootPaths(t *testing.T) {
	candidates, err := modelEndpointCandidates("https://models.example/api/v2")
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) < 4 || candidates[0] != "https://models.example/api/v2/v1/models" || candidates[1] != "https://models.example/api/v2/models" {
		t.Fatalf("unexpected model candidates: %#v", candidates)
	}
}

func TestParseModelListSortsAndDeduplicates(t *testing.T) {
	models, err := parseModelList([]byte(`{"data":[{"id":"zeta","owned_by":"test"},{"id":"alpha"},{"id":"zeta"},{"id":" "}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 2 || models[0].ID != "alpha" || models[1].ID != "zeta" || models[1].OwnedBy != "test" {
		t.Fatalf("unexpected models: %#v", models)
	}
}

func TestFetchProfileModelsFallsBackFromV1Models(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer sk-test-placeholder" {
			t.Errorf("authorization header = %q", request.Header.Get("Authorization"))
		}
		if request.URL.Path == "/v1/models" {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		if request.URL.Path != "/models" {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = writer.Write([]byte(`{"data":[{"id":"model-b"},{"id":"model-a","owned_by":"fixture"}]}`))
	}))
	defer server.Close()
	directory := t.TempDir()
	store := config.NewStore(filepath.Join(directory, "config.json"))
	cfg := config.Default(8765)
	cfg.Network.Mode = "direct"
	if err := store.Save(cfg); err != nil {
		t.Fatal(err)
	}
	service := NewDesktopService(newTestRuntime(t, directory, store, cfg))
	models, err := service.FetchProfileModels(ProfileInput{BaseURL: server.URL, APIKey: "sk-test-placeholder"})
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 2 || models[0].ID != "model-a" || models[1].ID != "model-b" {
		t.Fatalf("unexpected fetched models: %#v", models)
	}
}
