//go:build windows

/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Windows GitHub Release 版本检测、校验、替换和重启
 * @File          : Windows 桌面更新实现
 */
package desktop

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"codexrelay/internal/network"
	"codexrelay/internal/relay"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/updater"
	githubprovider "github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

const (
	updateRepository    = "xxloocee/CodexRelay"
	updateChecksumAsset = "SHA256SUMS"
)

func configureUpdater(app *application.App, relayRuntime *relay.Runtime) error {
	provider, err := githubprovider.New(githubprovider.Config{
		Repository:    updateRepository,
		ChecksumAsset: updateChecksumAsset,
		AssetMatcher:  windowsUpdateAssetMatcher,
		HTTPClient: &http.Client{
			Timeout:   15 * time.Minute,
			Transport: updateRoundTripper{runtime: relayRuntime},
		},
	})
	if err != nil {
		return err
	}
	return app.Updater.Init(updater.Config{
		CurrentVersion: applicationVersion,
		Providers:      []updater.Provider{provider},
		Window:         updater.WindowNone,
	})
}

type updateRoundTripper struct {
	runtime *relay.Runtime
}

func (transport updateRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	state := transport.runtime.State()
	if state == nil {
		return nil, errors.New("程序尚未初始化")
	}
	current, err := network.BuildTransport(state.Config.Network, network.DetectSystemProxy(), state.Config.ProxyPort)
	if err != nil {
		return nil, err
	}
	defer current.CloseIdleConnections()
	return current.RoundTrip(request)
}

func updatesSupported() bool { return true }

// CheckForUpdate 查询官方仓库最新稳定版本。后台检查只返回状态，不下载或替换程序。
func (s *DesktopService) CheckForUpdate() (UpdateInfo, error) {
	s.updateMu.Lock()
	defer s.updateMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()
	return checkForUpdate(ctx)
}

// InstallUpdate 重新确认最新版本后下载并校验 EXE，再交给 Wails helper 原子替换并重启。
func (s *DesktopService) InstallUpdate() error {
	s.updateMu.Lock()
	defer s.updateMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	info, err := checkForUpdate(ctx)
	if err != nil {
		return err
	}
	if !info.Available {
		return errors.New("当前已经是最新版本")
	}
	app := application.Get()
	if app == nil || app.Updater == nil {
		return errors.New("Windows 更新服务尚未初始化")
	}
	if err := app.Updater.DownloadAndInstall(ctx); err != nil {
		return fmt.Errorf("下载并校验更新失败: %w", err)
	}
	if err := app.Updater.Restart(context.Background()); err != nil {
		return fmt.Errorf("启动更新替换程序失败: %w", err)
	}
	return nil
}

func checkForUpdate(ctx context.Context) (UpdateInfo, error) {
	info := UpdateInfo{Supported: true, CurrentVersion: applicationVersion}
	app := application.Get()
	if app == nil || app.Updater == nil {
		return info, errors.New("Windows 更新服务尚未初始化")
	}
	release, err := app.Updater.Check(ctx)
	if err != nil {
		return info, fmt.Errorf("检查 GitHub 最新版本失败: %w", err)
	}
	if release == nil {
		return info, nil
	}
	if err := validateWindowsUpdate(release); err != nil {
		return info, err
	}
	info.Available = true
	info.LatestVersion = release.Version
	info.Name = release.Name
	info.Size = release.Artifact.Size
	if !release.PublishedAt.IsZero() {
		info.PublishedAt = release.PublishedAt.Format(time.RFC3339)
	}
	return info, nil
}

func validateWindowsUpdate(release *updater.Release) error {
	if release.Artifact.Platform != "windows" {
		return errors.New("最新版本没有匹配的 Windows 更新文件")
	}
	if release.Verification == nil || release.Verification.DigestAlgo != "sha256" || len(release.Verification.Digest) != sha256.Size {
		return fmt.Errorf("版本 %s 缺少有效的 SHA-256 校验信息，已拒绝更新", release.Version)
	}
	return nil
}

func windowsUpdateAssetMatcher(req updater.CheckRequest, assets []githubprovider.ReleaseAsset) int {
	arch := strings.ToLower(req.Arch)
	if arch != "amd64" && arch != "arm64" {
		return -1
	}
	suffix := "-" + arch + ".exe"
	for index, asset := range assets {
		name := strings.ToLower(asset.Name)
		if strings.HasPrefix(name, "codexrelay-") && strings.HasSuffix(name, suffix) && !strings.Contains(name, "-setup") {
			return index
		}
	}
	return -1
}
