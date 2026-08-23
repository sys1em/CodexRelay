/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : 外部客户端 JSON、JSON5 配置读写辅助
 * @File          : JSON 配置辅助
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package clientconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	json5 "github.com/titanous/json5"
)

func configureJSONEnv(path, endpoint, key, baseKey, tokenKey string) error {
	if err := backupClientFile(path); err != nil {
		return err
	}
	value := map[string]any{}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &value); err != nil {
			return fmt.Errorf("解析 %s 失败，请手动配置: %w", filepath.Base(path), err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("读取 %s 失败: %w", filepath.Base(path), err)
	}
	env, ok := value["env"].(map[string]any)
	if !ok {
		env = map[string]any{}
		value["env"] = env
	}
	env[baseKey] = endpoint
	env[tokenKey] = key
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("编码 %s 失败: %w", filepath.Base(path), err)
	}
	return writeClientFile(path, append(data, '\n'))
}

// readJSONObject 同时接受 JSON 和客户端实际使用的 JSON5 配置；不存在时返回空对象。
func readJSONObject(path string) (map[string]any, error) {
	value := map[string]any{}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return value, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &value); err != nil {
		value = map[string]any{}
		if json5Err := json5.Unmarshal(data, &value); json5Err != nil {
			return nil, fmt.Errorf("JSON: %v; JSON5: %w", err, json5Err)
		}
	}
	return value, nil
}

func writeJSONObject(path string, value map[string]any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeClientFile(path, append(data, '\n'))
}
