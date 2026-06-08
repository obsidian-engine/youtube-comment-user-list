export interface ParseKeywordsResult {
  keywords: string[]
  invalid: string[]
}

// 1行1ワードの txt を string[] に変換。
// CRLF/LF 双方対応、前後空白除去、空行除外、重複は先勝ちで排除。
// `,` を含む行は API のクエリ仕様上分裂するため除外して invalid に積む（呼び出し側が警告表示）。
export function parseKeywordsTxt(text: string): ParseKeywordsResult {
  const keywords: string[] = []
  const invalid: string[] = []
  const seen = new Set<string>()
  for (const raw of text.split(/\r?\n/)) {
    const line = raw.trim()
    if (!line) continue
    if (line.includes(',')) {
      invalid.push(line)
      continue
    }
    if (seen.has(line)) continue
    seen.add(line)
    keywords.push(line)
  }
  return { keywords, invalid }
}
