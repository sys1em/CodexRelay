/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : 原子 JSON 存储回归测试
 */
package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestReadJSONDoesNotOverwriteDamagedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	original := []byte(`{"broken":`)
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}
	var target map[string]any
	if _, err := ReadJSON(path, &target); err == nil {
		t.Fatal("损坏的 JSON 必须返回错误")
	}
	current, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(current, original) {
		t.Fatal("读取损坏 JSON 不应改写原文件")
	}
}

func TestWriteJSONAtomicRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	if err := WriteJSONAtomic(path, ".data-*.tmp", map[string]int{"value": 7}); err != nil {
		t.Fatal(err)
	}
	var target map[string]int
	exists, err := ReadJSON(path, &target)
	if err != nil || !exists || target["value"] != 7 {
		t.Fatalf("round trip = %v, %v, %+v", exists, err, target)
	}
}
