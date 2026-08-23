/*
 * @Author        : 顾青离
 * @Url            : sucaijun.com
 * @Email          : Ricky@LiHai.La
 * @Project        : CodexRelay
 * @Description    : 公告、工具栏、加载反馈和弹窗的本地界面冒烟测试
 * @File           : Playwright 前端脱敏状态检查
 * @Read me        : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind         : 二次开发请保留原版权信息，谢谢。
 */
const fs = require("fs");
const http = require("http");
const path = require("path");
const { chromium } = require("playwright");

const projectRoot = path.resolve(__dirname, "..");
const frontendRoot = path.join(projectRoot, "frontend");
const state = {
  version: "0.5.0",
  needsOnboarding: false,
  proxyPort: 8765,
  proxyUrl: "http://127.0.0.1:8765/codex",
  proxyUrls: { codex: "http://127.0.0.1:8765/codex" },
  localAccessToken: "sk-local-placeholder",
  dataDirectory: "C:\\Users\\Ricky-Desktop\\.CodexRelay",
  activeProfiles: {},
  profiles: [],
  clientConfigs: [
    { category: "codex", label: "Codex", configDir: "C:\\Users\\Ricky-Desktop\\.codex", configFile: "config.toml", skipConfigReplacement: false, status: "not_detected", statusText: "未检测到配置" },
    { category: "claude", label: "Claude", configDir: "C:\\Users\\Ricky-Desktop\\.claude", configFile: "settings.json", skipConfigReplacement: true, status: "not_detected", statusText: "未检测到配置" },
  ],
  network: { mode: "direct", proxyUrl: "" },
  systemProxy: { enabled: false, note: "" },
  requests: [],
  usage: { updatedAt: "", total: {}, profiles: {}, days: {} },
  uptimeSeconds: 1,
  preferences: { closeToTray: true, launchAtStartup: false, startHidden: false, defaultSource: "doge", defaultCategory: "codex", restoreViewMode: "current" },
  doge: {
    bound: false,
    walletUsd: 0,
    subscriptionsUsd: 0,
    totalUsd: 0,
    subscriptions: [],
    redemptionEnabled: false,
    topupLink: "",
    user: {},
    account: { userId: 7, nickname: "测试用户", email: "user@example.test", balanceUsd: 3, usedUsd: 0.5, requestCount: 12 },
    groups: [],
    tokens: [],
    notifications: {
      initialized: true,
      enabled: true,
      currentNotice: "# 当前公告\n\n**当前公告内容**\n\n[官方网站](https://example.com) [危险链接](javascript:alert(1))",
      announcements: [{ id: 11, content: "## 历史公告\n\n**历史公告内容**", publishDate: "2026-08-22T00:00:00Z", type: "default", read: false }],
      unreadCount: 1,
      alerts: [],
      lastSyncAt: "2026-08-22T00:00:00Z",
      lastSyncError: "",
    },
    syncIntervalMinutes: 3,
    lastSyncAt: "",
    lastSyncError: "",
  },
};
state.doge.notifications.alerts = [{ kind: "announcement", key: "announcement:11", title: "新的系统公告", message: "平台发布了新的公告", announcementId: 11 }];

function contentType(filePath) {
  return filePath.endsWith(".html") ? "text/html" : filePath.endsWith(".css") ? "text/css" : filePath.endsWith(".js") ? "text/javascript" : filePath.endsWith(".svg") ? "image/svg+xml" : "application/octet-stream";
}

const server = http.createServer((request, response) => {
  let requestPath = decodeURIComponent((request.url || "/").split("?")[0]);
  if (requestPath === "/wails/runtime.js") {
    response.writeHead(200, { "Content-Type": "text/javascript" });
  response.end(`export const CancellablePromise = Promise; export const Create = { Any: x => x, Array: f => xs => (xs || []).map(f), Map: () => x => ({}), Nullable: f => x => x == null ? null : f(x), }; export const Events = { On: (name, callback) => { globalThis.__wailsEvents = globalThis.__wailsEvents || {}; globalThis.__wailsEvents[name] = callback; }, }; const wait = ms => new Promise(resolve => setTimeout(resolve, ms)); export const Call = { ByID: async (id, ...args) => { if (id === 3062805628) return globalThis.__relayState; if (id === 1536866851 || id === 2451788025 || id === 1654359792) { await wait(450); return; } if (id === 215903019) { globalThis.__relayState.doge.notifications.unreadCount = 0; return; } if (id === 93543395) { globalThis.__relayState.doge.tokenSwitch = null; return; } if (id === 3647092161) { globalThis.__relayState.doge.tokenSwitch = null; return; } if (id === 3707202435) { globalThis.__relayState.needsOnboarding = false; return; } if (id === 1746439210) { globalThis.__relayState.proxyPort = args[0]; return; } if (id === 3841417432) { globalThis.__relayState.doge.syncIntervalMinutes = args[0]; return; } if (id === 1477342492) return (args[0] || "C:\\\\Users\\\\Ricky-Desktop\\\\.CodexRelay") + "\\\\picked"; if (id === 2510123141) { const client = globalThis.__relayState.clientConfigs.find(item => item.category === args[0]); if (client) client.configDir = args[1]; return; } if (id === 2542801733) { const client = globalThis.__relayState.clientConfigs.find(item => item.category === args[0]); if (client) client.skipConfigReplacement = args[1]; return; } if (id === 3032635246) { globalThis.__relayState.dataDirectory = args[0]; return; } return; } };`);
    return;
  }
  const filePath = requestPath === "/logo.png" ? path.join(projectRoot, "logo.png") : path.join(frontendRoot, requestPath === "/" ? "index.html" : requestPath.slice(1));
  if (!filePath.startsWith(frontendRoot) && filePath !== path.join(projectRoot, "logo.png")) {
    response.writeHead(403); response.end(); return;
  }
  if (!fs.existsSync(filePath)) { response.writeHead(404); response.end(); return; }
  response.writeHead(200, { "Content-Type": contentType(filePath) });
  fs.createReadStream(filePath).pipe(response);
});

(async () => {
  await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
  const address = server.address();
  const browser = await chromium.launch({ headless: true, executablePath: "C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe" });
  try {
    const page = await browser.newPage({ viewport: { width: 1170, height: 724 }, deviceScaleFactor: 1 });
    await page.addInitScript((value) => { globalThis.__relayState = value; }, state);
    const errors = [];
    page.on("pageerror", (error) => errors.push(error.message));
    await page.goto(`http://127.0.0.1:${address.port}/`, { waitUntil: "networkidle" });
    const toolbar = await page.locator(".toolbar-actions > *").evaluateAll((items) => items.map((item) => item.id || item.className));
    const expected = ["dogeQuotaWrap", "openDogeTopup", "announcementWrap", "refreshDoge", "openSettings", "addProfile"];
    if (toolbar.join() !== expected.join()) throw new Error(`工具栏顺序错误: ${toolbar.join()}`);
    await page.click("#addProfile");
    await page.fill("#profileName", "未保存确认测试");
    await page.click("#editorBack");
    if (await page.locator("#confirmModal").isHidden() || await page.locator("#confirmModal .confirm-brand .panel-kicker").textContent() !== "CodexRelay" || !(await page.locator("#confirmMessage").textContent()).includes("当前修改尚未保存")) throw new Error("CodexRelay 自定义确认弹窗未显示");
    await page.screenshot({ path: path.join(projectRoot, ".tmp", "confirm-ui-smoke.png") });
    await page.click("#confirmCancel");
    if (await page.locator("#editorView").isHidden()) throw new Error("取消未保存返回确认后未留在编辑页");
    await page.click("#editorBack");
    await page.click("#confirmAccept");
    if (await page.locator("#profilesView").isHidden()) throw new Error("确认放弃未保存修改后未返回主页");
    await page.click("#addProfile");
    await page.locator(".advanced-fields summary").click();
    if (await page.locator("#modelManagerTitle").textContent() !== "模型管理" || await page.locator("#fetchModels").count() !== 1 || await page.locator("#addModel").count() !== 1) throw new Error("编辑页模型管理入口未显示");
    await page.click("#addModel");
    if (await page.locator("#modelRows .model-row").count() !== 1) throw new Error("手动添加模型行失败");
    await page.reload({ waitUntil: "networkidle" });
    if (await page.locator("#announcementBadge").textContent() !== "1") throw new Error("公告未读数字未显示");
    await page.evaluate(() => {
      globalThis.__relayState.doge.bound = true;
      globalThis.__relayState.doge.tokens = [{ id: 42, name: "后台更新令牌", maskedKey: "sk-****", status: 1, group: "余额低价组", groupDisplayName: "余额低价组", groupRatio: 0.02, category: "codex", imported: false, needsCategory: false, permitted: true, active: false }];
      globalThis.__wailsEvents["relay-state-changed"]();
    });
    await page.waitForTimeout(50);
    if (await page.locator(".doge-token-row").count() !== 1 || !(await page.locator(".doge-token-row").textContent()).includes("后台更新令牌")) throw new Error("后台状态事件未刷新令牌目录");
    await page.evaluate(() => {
      globalThis.__relayState.doge.tokens = Array.from({ length: 20 }, (_, index) => ({
        id: 42 + index,
        name: `滚动测试令牌 ${index + 1}`,
        maskedKey: "sk-****",
        status: 1,
        group: "余额低价组",
        groupDisplayName: "余额低价组",
        groupRatio: 0.02,
        category: "codex",
        imported: false,
        needsCategory: false,
        permitted: true,
        active: false,
      }));
      globalThis.__wailsEvents["relay-state-changed"]();
    });
    await page.waitForTimeout(50);
    const profileScrollLayout = await page.evaluate(() => {
      const view = document.querySelector("#profilesView");
      const toolbar = view.querySelector(".app-toolbar");
      const filters = view.querySelector(".profile-toolbar");
      const list = view.querySelector(".profile-list");
      const before = { toolbarTop: toolbar.getBoundingClientRect().top, filtersTop: filters.getBoundingClientRect().top };
      list.scrollTop = list.scrollHeight;
      const after = { toolbarTop: toolbar.getBoundingClientRect().top, filtersTop: filters.getBoundingClientRect().top };
      return {
        overflowY: getComputedStyle(list).overflowY,
        scrollable: list.scrollHeight > list.clientHeight,
        outerScrollable: document.documentElement.scrollHeight > document.documentElement.clientHeight || document.body.scrollHeight > document.body.clientHeight,
        before,
        after,
      };
    });
    if (profileScrollLayout.overflowY !== "auto" || !profileScrollLayout.scrollable || profileScrollLayout.outerScrollable) throw new Error(`主页列表滚动区域错误: ${JSON.stringify(profileScrollLayout)}`);
    if (profileScrollLayout.before.toolbarTop !== profileScrollLayout.after.toolbarTop || profileScrollLayout.before.filtersTop !== profileScrollLayout.after.filtersTop) throw new Error(`主页顶部区域随列表滚动: ${JSON.stringify(profileScrollLayout)}`);
    await page.click("#openAnnouncements");
    if (await page.locator("#announcementPanel").isHidden()) throw new Error("公告面板未打开");
    if (!(await page.locator("#announcementNoticePane").textContent()).includes("当前公告内容")) throw new Error("当前公告未显示");
    if (await page.locator("#announcementNoticePane h1").count() !== 1) throw new Error("Markdown 标题未渲染");
    if (await page.locator("#announcementNoticePane strong").count() !== 1) throw new Error("Markdown 加粗未渲染");
    if (await page.locator("#announcementNoticePane a").count() !== 1) throw new Error("危险链接未被过滤");
    if (await page.locator("#announcementNoticePane a").getAttribute("target") !== "_blank") throw new Error("公告链接未设置安全打开方式");
    await page.click("#refreshDoge");
    const refreshLoadingClass = await page.locator("#refreshDoge .icon").getAttribute("class");
    if (!refreshLoadingClass.split(" ").includes("icon-load") || !refreshLoadingClass.split(" ").includes("spin")) throw new Error("刷新按钮未使用统一加载图标");
    await page.waitForTimeout(900);
    const refreshDoneClass = await page.locator("#refreshDoge .icon").getAttribute("class");
    if (!refreshDoneClass.split(" ").includes("icon-refresh") || refreshDoneClass.split(" ").includes("spin")) throw new Error(`刷新按钮静态图标未恢复: ${refreshDoneClass}`);
    await page.click('[data-announcement-tab="timeline"]');
    if (!(await page.locator("#announcementTimelinePane").textContent()).includes("历史公告内容")) throw new Error("历史公告未显示");
    if (await page.locator("#announcementTimelinePane h2").count() !== 1) throw new Error("历史公告 Markdown 标题未渲染");
    await page.click("#closeAnnouncements");
    await page.click("#openSettings");
    if (await page.locator("#defaultSource").inputValue() !== "doge" || await page.locator("#defaultCategory").inputValue() !== "codex") throw new Error("默认来源或类别未显示为二狗子/Codex");
    await page.click('#settingsTabs button[data-tab="advanced"]');
    if (await page.locator("#dataDirectory").inputValue() !== "C:\\Users\\Ricky-Desktop\\.CodexRelay") throw new Error("CodexRelay 默认数据目录未显示");
    if (await page.locator(".client-config-row").count() !== 2) throw new Error("外部客户端路径行未显示");
    if ((await page.locator(".client-config-row").first().textContent()).includes("config.toml") || (await page.locator(".client-config-row").nth(1).textContent()).includes("settings.json")) throw new Error("外部客户端配置文件名仍显示");
    await page.locator('.client-config-row').first().locator('button').first().click();
    await page.waitForTimeout(50);
    if (await page.locator('.client-config-row').first().locator('input[type="text"]').inputValue() !== "C:\\Users\\Ricky-Desktop\\.codex\\picked") throw new Error("外部客户端路径选择未保存");
    const claudeSkip = page.locator('.client-config-row').nth(1).locator('input[type="checkbox"]');
    if (!(await claudeSkip.isChecked())) throw new Error("跳过配置文件替换状态未显示");
    await claudeSkip.uncheck();
    if (await claudeSkip.isChecked()) throw new Error("跳过配置文件替换取消未生效");
    await page.locator("#chooseDataDirectory").click();
    await page.waitForTimeout(50);
    if (await page.locator("#dataDirectory").inputValue() !== "C:\\Users\\Ricky-Desktop\\.CodexRelay\\picked") throw new Error("CodexRelay 数据目录选择未保存");
    await page.click('#settingsTabs button[data-tab="general"]');
    const settingsScrollLayout = await page.evaluate(() => {
      const view = document.querySelector("#settingsView");
      const toolbar = view.querySelector(".subpage-toolbar");
      const tabs = view.querySelector("#settingsTabs");
      const content = view.querySelector("#settingsContent");
      const before = {
        toolbarTop: toolbar.getBoundingClientRect().top,
        tabsTop: tabs.getBoundingClientRect().top,
      };
      content.scrollTop = content.scrollHeight;
      const after = {
        toolbarTop: toolbar.getBoundingClientRect().top,
        tabsTop: tabs.getBoundingClientRect().top,
      };
      return {
        contentRight: content.getBoundingClientRect().right,
        panelRight: view.querySelector(".settings-panel:not(.hidden)").getBoundingClientRect().right,
        viewOverflow: getComputedStyle(view).overflow,
        contentOverflowY: getComputedStyle(content).overflowY,
        contentScrollable: content.scrollHeight > content.clientHeight,
        outerScrollable: document.documentElement.scrollHeight > document.documentElement.clientHeight || document.body.scrollHeight > document.body.clientHeight,
        before,
        after,
      };
    });
    if (settingsScrollLayout.viewOverflow !== "hidden" || settingsScrollLayout.contentOverflowY !== "auto" || !settingsScrollLayout.contentScrollable || settingsScrollLayout.outerScrollable || settingsScrollLayout.contentRight - settingsScrollLayout.panelRight < 20) throw new Error(`设置页滚动区域错误: ${JSON.stringify(settingsScrollLayout)}`);
    if (settingsScrollLayout.before.toolbarTop !== settingsScrollLayout.after.toolbarTop || settingsScrollLayout.before.tabsTop !== settingsScrollLayout.after.tabsTop) throw new Error(`设置页顶部或左侧导航随内容滚动: ${JSON.stringify(settingsScrollLayout)}`);
    await page.screenshot({ path: path.join(projectRoot, ".tmp", "settings-scroll-ui-smoke.png") });
    await page.click('#settingsTabs button[data-tab="network"]');
    if (await page.locator("#proxyPort").inputValue() !== "8765") throw new Error("默认监听端口未显示");
    await page.fill("#proxyPort", "18766");
    await page.click("#saveProxyPort");
    if (await page.locator("#proxyPort").inputValue() !== "18766") throw new Error("监听端口保存失败");
    await page.screenshot({ path: path.join(projectRoot, ".tmp", "network-port-ui-smoke.png"), fullPage: true });
    await page.click('#settingsTabs button[data-tab="about"]');
    const about = await page.locator("#aboutPanel").textContent();
    if (!about.includes("Ricky") || !about.includes("ergouzi.life") || !about.includes("dashboard/overview")) throw new Error("关于页信息未更新");
    const connection = await browser.newPage({ viewport: { width: 1120, height: 780 }, deviceScaleFactor: 1 });
    await connection.addInitScript((value) => { globalThis.__relayState = value; }, {
      ...state,
      doge: { ...state.doge, bound: true, redemptionEnabled: true, lastSyncAt: "2026-08-22T00:00:00Z", account: { userId: 7, nickname: "测试用户", email: "user@example.test", balanceUsd: 3, usedUsd: 0.5, requestCount: 12 } },
    });
    await connection.goto(`http://127.0.0.1:${address.port}/`, { waitUntil: "networkidle" });
    await connection.click("#openSettings");
    await connection.click('#settingsTabs button[data-tab="connection"]');
    const accountText = await connection.locator("#dogeBoundView").textContent();
    if (!accountText.includes("7") || !accountText.includes("测试用户") || !accountText.includes("user@example.test") || !accountText.includes("$3.00") || !accountText.includes("$0.50") || !accountText.includes("12")) throw new Error("连接页账户信息未显示");
    if (await connection.locator("#dogeSyncInterval").inputValue() !== "3") throw new Error("自动同步默认间隔不是 3 分钟");
    await connection.selectOption("#dogeSyncInterval", "1");
    await connection.waitForTimeout(50);
    if (await connection.locator("#dogeSyncInterval").inputValue() !== "1") throw new Error("自动同步间隔保存后未保持选择值");
    const syncOptions = await connection.locator("#dogeSyncInterval option").evaluateAll((options) => options.map((option) => option.value));
    if (syncOptions.join(",") !== "1,3,5,10,15,30,60") throw new Error("自动同步间隔选项不完整");
    await connection.evaluate(() => { globalThis.__wailsEvents["relay-state-changed"](); });
    await connection.waitForTimeout(50);
    if (await connection.locator("#dogeSyncInterval").inputValue() !== "1") throw new Error("重新读取状态后自动同步间隔被重置");
    await connection.click("#settingsBack");
    await connection.click("#openSettings");
    await connection.click('#settingsTabs button[data-tab="connection"]');
    if (await connection.locator("#dogeSyncInterval").inputValue() !== "1") throw new Error("重新打开设置后自动同步间隔未保持");
    await connection.screenshot({ path: path.join(projectRoot, ".tmp", "connection-account-ui-smoke.png"), fullPage: true });

    // 兑换弹窗打开期间，定时 loadState 不得重新计算滚动条补偿或造成主界面跳宽。
    await connection.click("#settingsBack");
    await connection.evaluate(() => { document.body.style.minHeight = "200vh"; });
    await connection.click("#openDogeTopup");
    if (await connection.locator("#dogeTopupModal").isHidden()) throw new Error("兑换弹窗未打开");
    const modalBefore = await connection.evaluate(() => ({
      shellWidth: document.querySelector("#appShell").getBoundingClientRect().width,
      bodyPaddingRight: getComputedStyle(document.body).paddingRight,
      compensation: document.body.style.getPropertyValue("--modal-scrollbar-compensation"),
      locked: document.body.classList.contains("modal-open"),
    }));
    await connection.waitForTimeout(3500);
    const modalAfter = await connection.evaluate(() => ({
      shellWidth: document.querySelector("#appShell").getBoundingClientRect().width,
      bodyPaddingRight: getComputedStyle(document.body).paddingRight,
      compensation: document.body.style.getPropertyValue("--modal-scrollbar-compensation"),
      locked: document.body.classList.contains("modal-open"),
    }));
    if (!modalBefore.locked || !modalAfter.locked) throw new Error("兑换弹窗未保持页面滚动锁定");
    if (modalBefore.shellWidth !== modalAfter.shellWidth) throw new Error(`弹窗期间主界面跳宽: ${modalBefore.shellWidth} -> ${modalAfter.shellWidth}`);
    if (modalBefore.bodyPaddingRight !== modalAfter.bodyPaddingRight || modalBefore.compensation !== modalAfter.compensation) throw new Error("弹窗期间滚动条补偿发生变化");
    await connection.click("#closeDogeTopupModal");
    const modalClosed = await connection.evaluate(() => ({
      paddingRight: getComputedStyle(document.body).paddingRight,
      compensation: document.body.style.getPropertyValue("--modal-scrollbar-compensation"),
      locked: document.body.classList.contains("modal-open"),
    }));
    if (modalClosed.locked || modalClosed.compensation || modalClosed.paddingRight !== "0px") throw new Error("关闭兑换弹窗后滚动条补偿未清理");
    await connection.evaluate(() => { document.body.style.minHeight = ""; });
    await connection.click("#openDogeTopup");
    await connection.fill("#dogeTopupCode", "fake-redemption-code");
    await connection.click("#submitDogeTopup");
    const topupLoadingClass = await connection.locator("#submitDogeTopup .icon").getAttribute("class");
    if (!topupLoadingClass.split(" ").includes("icon-load") || !topupLoadingClass.split(" ").includes("spin")) throw new Error("兑换按钮未使用统一加载图标");
    await connection.waitForTimeout(900);
    await connection.close();
    const popup = await browser.newPage({ viewport: { width: 450, height: 320 }, deviceScaleFactor: 1 });
    await popup.addInitScript((value) => { globalThis.__relayState = value; }, {
      ...state,
      doge: {
        ...state.doge,
        notifications: {
          ...state.doge.notifications,
          announcements: [{ ...state.doge.notifications.announcements[0], content: "# 公告标题\n\n" + "长公告内容 ".repeat(100) }],
        },
      },
    });
    await popup.goto(`http://127.0.0.1:${address.port}/notification.html?kind=announcement`, { waitUntil: "networkidle" });
    if (await popup.locator(".notification-markdown").count() !== 1) throw new Error("右下角公告正文未显示");
    if (await popup.locator(".notification-markdown h1").count() !== 1) throw new Error("右下角 Markdown 标题未渲染");
    const announcementText = await popup.locator(".notification-markdown").textContent();
    if (announcementText.length <= 300 || !announcementText.includes("长公告内容 长公告内容")) throw new Error("右下角公告正文被截断");
    const announcementCard = await popup.locator("#notificationCard").boundingBox();
    const announcementViewport = popup.viewportSize();
    if (!announcementCard || !announcementViewport || Math.abs(announcementCard.x - 10) > 1 || Math.abs(announcementCard.y - 10) > 1 || Math.abs(announcementViewport.width - announcementCard.x - announcementCard.width - 10) > 1 || Math.abs(announcementViewport.height - announcementCard.y - announcementCard.height - 10) > 1) throw new Error(`公告窗口边距不一致: ${JSON.stringify({ announcementCard, announcementViewport })}`);
    const announcementLayout = await popup.evaluate(() => {
      const card = document.querySelector("#notificationCard");
      const heading = document.querySelector(".notification-heading");
      const content = document.querySelector("#notificationContent");
      const actions = document.querySelector("#notificationActions");
      return { card, heading, content, actions };
    });
    if (!announcementLayout.content || !announcementLayout.actions) throw new Error("通知窗口缺少固定内容区或操作区");
    const announcementScroll = await popup.locator("#notificationContent").evaluate((node) => ({
      overflowY: getComputedStyle(node).overflowY,
      scrollable: node.scrollHeight > node.clientHeight,
    }));
    if (announcementScroll.overflowY !== "auto" || !announcementScroll.scrollable) throw new Error(`通知正文区域未启用独立滚动: ${JSON.stringify(announcementScroll)}`);
    if (!(await popup.locator("#notificationCard").evaluate((node) => getComputedStyle(node).overflow === "hidden"))) throw new Error("通知卡片外层仍可滚动");
    await popup.screenshot({ path: path.join(projectRoot, ".tmp", "announcement-popup-simulated.png") });
    await popup.close();
    const switchPopup = await browser.newPage({ viewport: { width: 450, height: 320 }, deviceScaleFactor: 1 });
    await switchPopup.addInitScript((value) => { globalThis.__relayState = value; }, {
      ...state,
      doge: {
        ...state.doge,
        tokenSwitch: {
          key: "doge-profile|auth",
          category: "codex",
          failureKind: "auth",
          failureCount: 5,
          failureStatus: 403,
          currentTokenId: 41,
          currentName: "当前令牌",
          currentGroup: "余额低价组",
          currentRatio: 0.02,
          message: "当前令牌连续 5 次返回 HTTP 403，是否切换同类别的其他可用令牌？",
          candidates: [
            { tokenId: 42, name: "Codex 低价组", source: "二狗子 API", group: "余额低价组", ratio: 0.02, selectable: true },
            { tokenId: 43, name: "Codex 稳定组", source: "二狗子 API", group: "余额稳定组", ratio: 0.025, selectable: true },
          ],
        },
      },
    });
    await switchPopup.goto(`http://127.0.0.1:${address.port}/notification.html?kind=token-switch`, { waitUntil: "networkidle" });
    if (!(await switchPopup.locator("#tokenSwitchMessage").textContent()).includes("HTTP 403")) throw new Error("令牌切换提示未显示");
    const switchCandidates = await switchPopup.locator("#tokenSwitchCandidates option").allTextContents();
    if (switchCandidates.join("|") !== "Codex 低价组(余额低价组 · 0.02)|Codex 稳定组(余额稳定组 · 0.025)") throw new Error("候选令牌格式错误");
    if (await switchPopup.locator("#cancelTokenSwitch").textContent() !== "取消" || await switchPopup.locator("#confirmTokenSwitch").textContent() !== "确定") throw new Error("令牌切换按钮文案错误");
    if (await switchPopup.locator("button:visible").count() !== 2 || !(await switchPopup.locator("#dismissNotification").isHidden())) throw new Error("令牌切换窗口出现多余操作按钮");
    const switchCard = await switchPopup.locator("#notificationCard").boundingBox();
    const switchViewport = switchPopup.viewportSize();
    if (!switchCard || !switchViewport || Math.abs(switchCard.x - 10) > 1 || Math.abs(switchCard.y - 10) > 1 || Math.abs(switchViewport.width - switchCard.x - switchCard.width - 10) > 1 || Math.abs(switchViewport.height - switchCard.y - switchCard.height - 10) > 1) throw new Error(`令牌切换窗口边距不一致: ${JSON.stringify({ switchCard, switchViewport })}`);
    if (!announcementCard || Math.abs(announcementCard.width - switchCard.width) > 1 || Math.abs(announcementCard.height - switchCard.height) > 1) throw new Error(`提醒窗口尺寸未统一: ${JSON.stringify({ announcementCard, switchCard })}`);
    await switchPopup.screenshot({ path: path.join(projectRoot, ".tmp", "token-switch-ui-smoke.png") });
    await switchPopup.click("#cancelTokenSwitch");
    await switchPopup.close();
    for (const variant of [
      {
        status: 401,
        message: "当前令牌“Codex 低价组”连续 5 次返回 HTTP 401，是否切换同类别的其他可用令牌？",
        screenshot: "token-switch-401-ui-smoke.png",
      },
      {
        status: 502,
        message: "当前令牌“Codex 低价组”在 3 分钟内出现 5 次上游异常。上游连续异常，是否尝试切换同类别的其他可用令牌？",
        screenshot: "token-switch-5xx-ui-smoke.png",
      },
    ]) {
      const variantPopup = await browser.newPage({ viewport: { width: 450, height: 320 }, deviceScaleFactor: 1 });
      await variantPopup.addInitScript((value) => { globalThis.__relayState = value; }, {
        ...state,
        doge: {
          ...state.doge,
          tokenSwitch: {
            key: "doge-profile|variant",
            category: "codex",
            failureKind: variant.status === 401 ? "auth" : "upstream",
            failureCount: 5,
            failureStatus: variant.status,
            currentTokenId: 41,
            currentName: "Codex 低价组",
            currentGroup: "余额低价组",
            currentRatio: 0.02,
            message: variant.message,
            candidates: [{ tokenId: 42, name: "Codex 低价组", source: "二狗子 API", group: "余额低价组", ratio: 0.02, selectable: true }],
          },
        },
      });
      await variantPopup.goto(`http://127.0.0.1:${address.port}/notification.html?kind=token-switch`, { waitUntil: "networkidle" });
      if (!(await variantPopup.locator("#tokenSwitchMessage").textContent()).includes(variant.status === 401 ? "HTTP 401" : "3 分钟内出现 5 次上游异常")) throw new Error(`令牌切换 ${variant.status} 文案未显示`);
      await variantPopup.screenshot({ path: path.join(projectRoot, ".tmp", variant.screenshot) });
      await variantPopup.close();
    }
    const onboarding = await browser.newPage({ viewport: { width: 560, height: 700 }, deviceScaleFactor: 1 });
    await onboarding.addInitScript((value) => { globalThis.__relayState = value; }, { ...state, needsOnboarding: true });
    await onboarding.goto(`http://127.0.0.1:${address.port}/`, { waitUntil: "networkidle" });
    if (await onboarding.locator("#onboardingModal").isHidden()) throw new Error("首次引导弹窗未显示");
    const onboardingText = await onboarding.locator("#onboardingModal").textContent();
    if (!onboardingText.includes("打开二狗子用户中心 → https://ergouzi.life/profile")) throw new Error("用户中心地址未显示");
    if (!onboardingText.includes("安全 → 访问令牌")) throw new Error("首次引导路径未显示");
    await onboarding.click("#openDogeProfile");
    await onboarding.click("#bindOnboarding");
    if (await onboarding.locator("#onboardingError").isHidden()) throw new Error("空令牌校验未显示");
    await onboarding.screenshot({ path: path.join(projectRoot, ".tmp", "onboarding-ui-smoke.png"), fullPage: true });
    await onboarding.click("#skipOnboarding");
    if (!(await onboarding.locator("#onboardingModal").isHidden())) throw new Error("跳过首次引导后弹窗仍显示");
    await onboarding.close();
    await page.setViewportSize({ width: 390, height: 844 });
    await page.click('#settingsTabs button[data-tab="general"]');
    const mobileSettingsLayout = await page.evaluate(() => {
      const view = document.querySelector("#settingsView");
      const tabs = view.querySelector("#settingsTabs");
      const content = view.querySelector("#settingsContent");
      const before = tabs.getBoundingClientRect().top;
      content.scrollTop = content.scrollHeight;
      return {
        contentOverflowY: getComputedStyle(content).overflowY,
        contentScrollable: content.scrollHeight > content.clientHeight,
        tabsTop: Math.abs(before - tabs.getBoundingClientRect().top) < 1,
        outerScrollable: document.documentElement.scrollHeight > document.documentElement.clientHeight || document.body.scrollHeight > document.body.clientHeight,
      };
    });
    if (mobileSettingsLayout.contentOverflowY !== "auto" || !mobileSettingsLayout.contentScrollable || !mobileSettingsLayout.tabsTop || mobileSettingsLayout.outerScrollable) throw new Error(`移动端设置页滚动区域错误: ${JSON.stringify(mobileSettingsLayout)}`);
    const overflow = await page.evaluate(() => ({ width: document.documentElement.scrollWidth, viewport: document.documentElement.clientWidth }));
    if (overflow.width > overflow.viewport) throw new Error(`移动端横向溢出: ${JSON.stringify(overflow)}`);
    await page.screenshot({ path: path.join(projectRoot, ".tmp", "notification-ui-smoke.png"), fullPage: true });
    if (errors.length) throw new Error(`页面错误: ${errors.join("; ")}`);
    console.log(JSON.stringify({ toolbar, overflow, screenshot: ".tmp/notification-ui-smoke.png" }));
  } finally {
    await browser.close();
    server.close();
  }
})().catch((error) => { console.error(error); process.exitCode = 1; });
