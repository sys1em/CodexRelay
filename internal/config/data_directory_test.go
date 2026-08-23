/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : CodexRelay 数据目录指针回归测试
 * @File          : 数据目录定位测试
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDataDirectoryPointerRoundTrip(t *testing.T) {
	defaultDirectory := filepath.Join(t.TempDir(), ".CodexRelay")
	customDirectory := filepath.Join(t.TempDir(), "custom")
	if got, err := resolveDataDirectory(defaultDirectory); err != nil || got != defaultDirectory {
		t.Fatalf("default data directory = %q, err=%v", got, err)
	}
	if err := saveDataDirectoryPointer(defaultDirectory, customDirectory); err != nil {
		t.Fatal(err)
	}
	if got, err := resolveDataDirectory(defaultDirectory); err != nil || got != customDirectory {
		t.Fatalf("custom data directory = %q, err=%v", got, err)
	}
	if err := saveDataDirectoryPointer(defaultDirectory, defaultDirectory); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(defaultDirectory, dataDirectoryPointerFile)); !os.IsNotExist(err) {
		t.Fatalf("default pointer should be removed, stat err=%v", err)
	}
}

func TestDataDirectoryPointerRejectsRelativePath(t *testing.T) {
	defaultDirectory := filepath.Join(t.TempDir(), ".CodexRelay")
	if err := saveDataDirectoryPointer(defaultDirectory, "relative"); err == nil {
		t.Fatal("relative data directory should fail")
	}
}
