//go:build windows

/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Windows 外部 URL 校验回归测试
 * @File          : 默认浏览器 URL 校验测试
 */
package platform

import "testing"

func TestValidateExternalURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{name: "https", input: "https://example.test/shop", valid: true},
		{name: "http", input: "http://example.test/shop", valid: true},
		{name: "javascript", input: "javascript:alert(1)", valid: false},
		{name: "missing host", input: "https:///shop", valid: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := validateExternalURL(test.input)
			if (err == nil) != test.valid {
				t.Fatalf("validateExternalURL(%q) error = %v, valid = %v", test.input, err, test.valid)
			}
		})
	}
}
