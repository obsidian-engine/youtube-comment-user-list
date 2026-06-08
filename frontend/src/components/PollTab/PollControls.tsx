import { useRef } from 'react'
import { LoadingButton } from '../LoadingButton'

interface PollControlsProps {
  keywords: string[]
  onLoadFile: (file: File) => void | Promise<void>
  onClear: () => void
  onRecount: () => void
  isLoading: boolean
  lastUpdated: string
}

const SAMPLE_TXT = '賛成\n反対\n保留\n'

function downloadSampleTxt() {
  const blob = new Blob([SAMPLE_TXT], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'keywords-sample.txt'
  a.click()
  URL.revokeObjectURL(url)
}

export function PollControls({
  keywords,
  onLoadFile,
  onClear,
  onRecount,
  isLoading,
  lastUpdated,
}: PollControlsProps) {
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      await onLoadFile(file)
    }
    e.target.value = ''
  }

  return (
    <div className="space-y-4">
      <section className="rounded-lg shadow-subtle ring-1 ring-black/5 bg-white/80 backdrop-blur p-5">
        <h2 className="text-sm font-semibold mb-3 text-slate-700">
          投票キーワード（txt から読み込み）
        </h2>

        <p className="text-[12px] text-slate-500 mb-3">
          メモ帳で 1 行 1 ワードの txt
          を作成し、「読み込み」ボタンで取り込んでください。コメントが完全一致した場合のみ 1
          票としてカウントされます。
        </p>

        <input
          ref={fileInputRef}
          type="file"
          accept=".txt,text/plain"
          onChange={handleFileChange}
          className="hidden"
        />

        <div className="flex gap-2 mb-4">
          <LoadingButton
            onClick={() => fileInputRef.current?.click()}
            disabled={isLoading}
            variant="primary"
          >
            キーワードtxtを読み込み
          </LoadingButton>
          <LoadingButton onClick={downloadSampleTxt} disabled={isLoading} variant="outline">
            サンプルtxtをダウンロード
          </LoadingButton>
          {keywords.length > 0 && (
            <LoadingButton onClick={onClear} disabled={isLoading} variant="outline">
              クリア
            </LoadingButton>
          )}
        </div>

        <div className="flex flex-wrap gap-2">
          {keywords.length === 0 && (
            <span className="text-[12px] text-slate-500">
              キーワード未設定。txt を読み込むと一覧表示されます。
            </span>
          )}
          {keywords.map((word) => (
            <span
              key={word}
              className="inline-flex items-center px-3 py-1 rounded-full bg-slate-200 text-sm"
            >
              {word}
            </span>
          ))}
        </div>
      </section>

      <div className="flex items-center justify-between">
        <LoadingButton
          onClick={onRecount}
          isLoading={isLoading}
          loadingText="集計中..."
          variant="primary"
          disabled={keywords.length === 0}
        >
          今すぐ集計
        </LoadingButton>
        <div className="text-[12px] text-slate-500">最終更新: {lastUpdated}</div>
      </div>
    </div>
  )
}
