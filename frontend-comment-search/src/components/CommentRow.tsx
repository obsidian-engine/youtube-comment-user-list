import type { Comment } from '../types'

interface CommentRowProps {
  comment: Comment
  index: number
  isChecked: boolean
  onToggle: () => void
}

export function CommentRow({ comment, index, isChecked, onToggle }: CommentRowProps) {
  const time = new Date(comment.publishedAt).toLocaleTimeString('ja-JP', {
    hour: '2-digit',
    minute: '2-digit',
  })

  return (
    <tr
      className={`transition-colors duration-150 hover:bg-slate-200/40 dark:hover:bg-slate-700/20 ${
        isChecked ? 'opacity-50 bg-green-50 dark:bg-green-900/20' : ''
      } ${
        index % 2 === 0
          ? 'bg-slate-100/50 dark:bg-slate-800/20'
          : 'bg-slate-200/40 dark:bg-slate-700/25'
      }`}
    >
      <td className="px-4 py-3 text-center">
        <input
          type="checkbox"
          checked={isChecked}
          onChange={onToggle}
          className="w-5 h-5 rounded cursor-pointer"
          aria-label="読み上げ済み"
        />
      </td>
      <td className="px-4 py-3 font-mono text-[13px] text-center">{time}</td>
      <td className="px-4 py-3 font-medium">{comment.displayName}</td>
      <td className="px-4 py-3">{comment.message}</td>
    </tr>
  )
}
