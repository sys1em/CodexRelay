/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : 便携 JSON 文件共享原子读写
 */
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ReadJSON 区分文件不存在和文件损坏，调用方据此决定是否创建默认数据。
func ReadJSON(path string, target any) (bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("读取 %s: %w", filepath.Base(path), err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return true, fmt.Errorf("解析 %s: %w", filepath.Base(path), err)
	}
	return true, nil
}

// WriteJSONAtomic 在目标目录内写临时文件并同步后替换，避免进程中断留下半截 JSON。
func WriteJSONAtomic(path, temporaryPattern string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("编码 JSON: %w", err)
	}
	data = append(data, '\n')
	return WriteBytesAtomic(path, temporaryPattern, data, 0o600)
}

// WriteBytesAtomic 在目标目录内原子替换任意配置文件。
// 外部客户端配置必须先由调用方创建备份，再通过此函数写入，避免进程中断留下半截文件。
func WriteBytesAtomic(path, temporaryPattern string, data []byte, mode os.FileMode) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("创建数据目录: %w", err)
	}

	temporary, err := os.CreateTemp(directory, temporaryPattern)
	if err != nil {
		return fmt.Errorf("创建临时文件: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	closeWithError := func(operationErr error) error {
		_ = temporary.Close()
		return operationErr
	}
	if mode == 0 {
		mode = 0o600
	}
	if err := temporary.Chmod(mode); err != nil {
		return closeWithError(err)
	}
	if _, err := temporary.Write(data); err != nil {
		return closeWithError(fmt.Errorf("写入临时文件: %w", err))
	}
	if err := temporary.Sync(); err != nil {
		return closeWithError(fmt.Errorf("同步临时文件: %w", err))
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := replaceFile(temporaryName, path); err != nil {
		return fmt.Errorf("替换数据文件: %w", err)
	}
	return nil
}
