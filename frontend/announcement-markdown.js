/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : 公告 Markdown 解析与安全过滤
 * @File          : 主窗口和提醒窗口共用的公告正文渲染器
 * @Read me       : 感谢使用 CodexRelay，源码注释齐全，支持二次开发。
 * @Remind        : 二次开发请保留原版权信息，谢谢。
 */
import { marked } from "./vendor/marked.esm.js";

const allowedTags = new Set([
  "DIV", "P", "BR", "STRONG", "B", "EM", "I", "DEL", "S",
  "UL", "OL", "LI", "SPAN", "H1", "H2", "H3", "H4", "BLOCKQUOTE",
  "A", "PRE", "CODE", "HR", "TABLE", "THEAD", "TBODY", "TR", "TH", "TD",
]);

function sanitizeAnnouncementHTML(value) {
  const source = document.createElement("div");
  source.innerHTML = String(value || "");

  const copy = (node) => {
    if (node.nodeType === Node.TEXT_NODE) return document.createTextNode(node.textContent || "");
    if (node.nodeType !== Node.ELEMENT_NODE || !allowedTags.has(node.tagName)) {
      const fragment = document.createDocumentFragment();
      for (const child of node.childNodes || []) fragment.append(copy(child));
      return fragment;
    }
    const result = document.createElement(node.tagName.toLowerCase());
    if (node.tagName === "A") {
      const href = node.getAttribute("href") || "";
      if (!/^https?:\/\//i.test(href)) {
        const fragment = document.createDocumentFragment();
        for (const child of node.childNodes || []) fragment.append(copy(child));
        return fragment;
      }
      result.setAttribute("href", href);
      result.setAttribute("target", "_blank");
      result.setAttribute("rel", "noreferrer noopener");
    }
    for (const child of node.childNodes) result.append(copy(child));
    return result;
  };

  const result = document.createDocumentFragment();
  for (const child of source.childNodes) result.append(copy(child));
  return result;
}

// 将原始公告转换为安全 DOM；正文长度由通知窗口内部滚动承载，不在渲染阶段截断。
export function renderAnnouncementMarkdown(value) {
  let html;
  try {
    html = marked.parse(String(value || ""), {
      breaks: true,
      gfm: true,
      headerIds: false,
      html: true,
      mangle: false,
    });
  } catch {
    html = String(value || "");
  }
  return sanitizeAnnouncementHTML(html);
}
