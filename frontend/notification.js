/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : 右下角提醒窗口状态读取与确认交互
 * @File          : 独立提醒窗口脚本
 */
import { DismissDogeNotification, DismissDogeTokenSwitch, GetState, SwitchDogeToken } from "./api.js";
import { renderAnnouncementMarkdown } from "./announcement-markdown.js";
import * as wails from "/wails/runtime.js";

const kind = new URLSearchParams(window.location.search).get("kind") || "announcement";
const $ = (id) => document.getElementById(id);
let switchPrompt = null;

function setButtonLoading(button, loading, label = "") {
  if (!button) return;
  let iconNode = button.querySelector(".icon");
  const labelNode = button.querySelector("[data-button-label]");
  if (loading) {
    if (button.dataset.loading !== "true") {
      button.dataset.loading = "true";
      if (!iconNode) {
        iconNode = document.createElement("span");
        iconNode.className = "icon icon-load spin";
        iconNode.dataset.loadingCreated = "true";
        button.prepend(iconNode);
      }
      button.dataset.loadingIcon = iconNode.className;
      if (labelNode) button.dataset.loadingLabel = labelNode.textContent;
    }
    button.disabled = true;
    iconNode = button.querySelector(".icon");
    if (iconNode) iconNode.className = "icon icon-load spin";
    if (labelNode && label) labelNode.textContent = label;
    return;
  }
  if (button.dataset.loading !== "true") return;
  if (iconNode?.dataset.loadingCreated === "true") iconNode.remove();
  else if (iconNode) iconNode.className = button.dataset.loadingIcon || "icon icon-load";
  if (labelNode) labelNode.textContent = button.dataset.loadingLabel || labelNode.textContent;
  delete button.dataset.loading;
  delete button.dataset.loadingIcon;
  delete button.dataset.loadingLabel;
  button.disabled = false;
}

function formatRatio(value) {
  const ratio = Number(value);
  if (!Number.isFinite(ratio) || ratio <= 0) return "未知";
  return ratio.toFixed(4).replace(/0+$/, "").replace(/\.$/, "");
}

function renderTokenSwitch(prompt) {
  switchPrompt = prompt || null;
  const panel = $("tokenSwitchPanel");
  const rows = $("notificationRows");
  const dismiss = $("dismissNotification");
  const actions = $("tokenSwitchActions");
  const title = $("notificationTitle");
  panel.hidden = !prompt;
  rows.hidden = true;
  dismiss.hidden = true;
  actions.hidden = !prompt;
  title.textContent = "令牌切换提醒";
  const select = $("tokenSwitchCandidates");
  const status = $("tokenSwitchStatus");
  const message = $("tokenSwitchMessage");
  select.replaceChildren();
  message.textContent = prompt?.message || "当前令牌连续异常，建议尝试切换其他令牌。";
  status.textContent = "";
  const candidates = (prompt?.candidates || []).filter((candidate) => candidate.selectable !== false);
  for (const candidate of candidates) {
    const option = document.createElement("option");
    option.value = String(candidate.tokenId);
    const name = candidate.name || `令牌 ${candidate.tokenId}`;
    const group = candidate.group || "未知分组";
    option.textContent = `${name}(${group} · ${formatRatio(candidate.ratio)})`;
    select.appendChild(option);
  }
  select.disabled = candidates.length === 0;
  $("confirmTokenSwitch").disabled = candidates.length === 0;
  if (prompt && candidates.length === 0) {
    status.textContent = "当前类别没有可用的其他令牌";
  }
}

function render(state) {
  if (kind === "token-switch") {
    renderTokenSwitch(state?.doge?.tokenSwitch || null);
    return;
  }
  switchPrompt = null;
  $("tokenSwitchPanel").hidden = true;
  $("notificationRows").hidden = false;
  $("dismissNotification").hidden = false;
  $("tokenSwitchActions").hidden = true;
  const alerts = state?.doge?.notifications?.alerts || [];
  const filtered = alerts.filter((alert) => alert.kind === kind);
  const labels = { balance: "余额提醒", subscription: "套餐提醒", announcement: "系统公告" };
  $("notificationTitle").textContent = labels[kind] || "CodexRelay 提醒";
  const rows = $("notificationRows");
  rows.replaceChildren();
  for (const alert of filtered) {
    const row = document.createElement("div");
    row.className = `notification-row ${kind}`;
    const title = document.createElement("strong");
    title.textContent = alert.title || labels[kind] || "提醒";
    const announcement = kind === "announcement"
      ? (state?.doge?.notifications?.announcements || []).find((item) => Number(item.id) === Number(alert.announcementId))
      : null;
    const message = document.createElement(kind === "announcement" ? "div" : "span");
    if (announcement) {
      message.className = "notification-markdown";
      message.append(renderAnnouncementMarkdown(announcement.content || alert.message));
    } else {
      message.textContent = alert.message || "请查看账户状态";
    }
    row.append(title, message);
    rows.appendChild(row);
  }
}

$("cancelTokenSwitch").addEventListener("click", async () => {
  if (!switchPrompt?.key) return;
  const button = $("cancelTokenSwitch");
  setButtonLoading(button, true, "取消中...");
  $("confirmTokenSwitch").disabled = true;
  try {
    await DismissDogeTokenSwitch(switchPrompt.key);
  } catch (error) {
    $("tokenSwitchStatus").textContent = error?.message || String(error || "取消失败");
    setButtonLoading(button, false);
    $("confirmTokenSwitch").disabled = false;
  }
});

$("confirmTokenSwitch").addEventListener("click", async () => {
  if (!switchPrompt?.key) return;
  const tokenID = Number($("tokenSwitchCandidates").value);
  if (!Number.isSafeInteger(tokenID) || tokenID <= 0) return;
  const button = $("confirmTokenSwitch");
  setButtonLoading(button, true, "切换中...");
  $("cancelTokenSwitch").disabled = true;
  $("tokenSwitchStatus").textContent = "切换中...";
  try {
    await SwitchDogeToken(switchPrompt.key, tokenID);
    await loadState();
  } catch (error) {
    $("tokenSwitchStatus").textContent = error?.message || String(error || "切换失败");
    setButtonLoading(button, false);
    $("cancelTokenSwitch").disabled = false;
  }
});

async function loadState() {
  try { render(await GetState()); } catch { /* 主窗口会继续负责显示同步错误。 */ }
}

$("dismissNotification").addEventListener("click", async () => {
  const button = $("dismissNotification");
  setButtonLoading(button, true, "关闭中...");
  try { await DismissDogeNotification(kind); } finally { setButtonLoading(button, false); }
});
wails.Events.On("notification-state-changed", loadState);
loadState();
