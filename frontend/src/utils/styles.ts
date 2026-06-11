// 共通 utility class 定義 (ネイティブ要素の OS 既定スタイル上書き用)。

export const SELECT_CLASS =
  'appearance-none cursor-pointer text-[12px] pl-2 pr-7 py-1 ' +
  'bg-c-bg-2 border border-c-line-strong ' +
  'text-c-ink font-mono ' +
  'focus:outline-none focus:ring-2 focus:ring-c-accent-2/40 ' +
  'bg-no-repeat bg-[length:14px_14px] bg-[position:right_0.4rem_center] ' +
  "bg-[image:url('data:image/svg+xml;charset=utf-8,%3Csvg%20xmlns=%22http://www.w3.org/2000/svg%22%20viewBox=%220%200%2024%2024%22%20fill=%22none%22%20stroke=%22%230a0a0f%22%20stroke-width=%222%22%20stroke-linecap=%22round%22%20stroke-linejoin=%22round%22%3E%3Cpolyline%20points=%226%209%2012%2015%2018%209%22/%3E%3C/svg%3E')] " +
  'disabled:opacity-50 disabled:cursor-not-allowed'
