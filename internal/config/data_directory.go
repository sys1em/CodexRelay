/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : CodexRelay 便携数据目录定位与路径指针
 * @File          : CodexRelay 数据目录定位
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"codexrelay/internal/storage"
)

const dataDirectoryPointerFile = ".active-directory.json"

type dataDirectoryPointer struct {
	Directory string `json:"directory"`
}

// DefaultDataDirectory 返回当前用户的默认 CodexRelay 数据目录；程序运行目录不参与定位。
func DefaultDataDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("读取用户目录失败: %w", err)
	}
	home = strings.TrimSpace(home)
	if home == "" {
		return "", errors.New("用户目录为空")
	}
	return filepath.Join(home, ".CodexRelay"), nil
}

// ResolveDataDirectory 只从默认目录中的路径指针解析当前数据目录；指针损坏时明确失败，不回退到未知目录。
func ResolveDataDirectory() (string, error) {
	defaultDirectory, err := DefaultDataDirectory()
	if err != nil {
		return "", err
	}
	return resolveDataDirectory(defaultDirectory)
}

func resolveDataDirectory(defaultDirectory string) (string, error) {
	pointerPath := filepath.Join(defaultDirectory, dataDirectoryPointerFile)
	var pointer dataDirectoryPointer
	exists, err := storage.ReadJSON(pointerPath, &pointer)
	if err != nil {
		return "", fmt.Errorf("读取 CodexRelay 数据目录指针失败: %w", err)
	}
	if !exists {
		return defaultDirectory, nil
	}
	directory := strings.TrimSpace(pointer.Directory)
	if directory == "" || !filepath.IsAbs(directory) {
		return "", fmt.Errorf("CodexRelay 数据目录指针不是绝对路径: %q", pointer.Directory)
	}
	return filepath.Clean(directory), nil
}

// SaveDataDirectoryPointer 原子保存当前自定义目录；切回默认目录时删除指针以恢复默认定位。
func SaveDataDirectoryPointer(directory string) error {
	defaultDirectory, err := DefaultDataDirectory()
	if err != nil {
		return err
	}
	return saveDataDirectoryPointer(defaultDirectory, directory)
}

func saveDataDirectoryPointer(defaultDirectory, directory string) error {
	directory = filepath.Clean(strings.TrimSpace(directory))
	if directory == "" || !filepath.IsAbs(directory) {
		return errors.New("CodexRelay 数据目录必须是绝对路径")
	}
	pointerPath := filepath.Join(defaultDirectory, dataDirectoryPointerFile)
	if directory == filepath.Clean(defaultDirectory) {
		if err := os.Remove(pointerPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("清理 CodexRelay 数据目录指针失败: %w", err)
		}
		return nil
	}
	if err := storage.WriteJSONAtomic(pointerPath, ".active-directory-*.tmp", dataDirectoryPointer{Directory: directory}); err != nil {
		return fmt.Errorf("保存 CodexRelay 数据目录指针失败: %w", err)
	}
	return nil
}
