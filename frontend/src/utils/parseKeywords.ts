export function parseKeywords(input: string): string[] {
  return [
    ...new Set(
      input
        .split(/[\n,]/)
        .map((k) => k.trim())
        .filter((k) => k.length > 0),
    ),
  ]
}
