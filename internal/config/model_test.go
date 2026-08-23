/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : 本地模型目录校验回归测试
 * @File          : 模型配置测试
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
package config

import "testing"

func TestValidateModelsRequiresUniqueDefault(t *testing.T) {
	models := []ModelEntry{{ID: "model-a"}, {ID: "model-b", Name: "B"}}
	if err := ValidateModels(models, "model-b"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateModels(models, "missing"); err == nil {
		t.Fatal("missing default model should fail")
	}
	if err := ValidateModels([]ModelEntry{{ID: "model-a"}, {ID: "model-a"}}, "model-a"); err == nil {
		t.Fatal("duplicate model ID should fail")
	}
}
