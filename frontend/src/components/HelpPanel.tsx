import { useState, type ReactNode } from 'react'

interface SectionProps {
  id: string
  title: string
  defaultOpen?: boolean
  children: ReactNode
}

function Section({ id, title, defaultOpen = false, children }: SectionProps) {
  const [open, setOpen] = useState(defaultOpen)
  return (
    <section
      style={{
        border: '1px solid var(--c-line-strong)',
        background: 'var(--c-bg-2)',
        overflow: 'hidden',
      }}
    >
      <button
        type="button"
        onClick={() => setOpen(!open)}
        aria-expanded={open}
        aria-controls={`help-${id}`}
        style={{
          width: '100%',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: '12px',
          padding: '14px 16px',
          textAlign: 'left',
          background: 'none',
          border: 'none',
          cursor: 'pointer',
          borderBottom: open ? '1px solid var(--c-line)' : 'none',
        }}
      >
        <span
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            fontFamily: 'var(--f-body)',
            fontWeight: 700,
            fontSize: '14px',
            color: 'var(--c-ink)',
          }}
        >
          <svg
            style={{
              width: '14px',
              height: '14px',
              transition: 'transform 0.2s',
              transform: open ? 'rotate(90deg)' : 'none',
              color: 'var(--c-accent-2)',
            }}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
          </svg>
          {title}
        </span>
      </button>
      {open && (
        <div
          id={`help-${id}`}
          style={{
            padding: '16px',
            fontSize: '13px',
            lineHeight: 1.75,
            color: 'var(--c-ink-dim)',
          }}
          className="space-y-3"
        >
          {children}
        </div>
      )}
    </section>
  )
}

function Note({
  tone = 'info',
  children,
}: {
  tone?: 'info' | 'warn' | 'danger'
  children: ReactNode
}) {
  const borderColor: Record<string, string> = {
    info: 'var(--c-accent-2)',
    warn: '#b56b00',
    danger: 'var(--c-error)',
  }
  const labels: Record<string, string> = {
    info: 'ポイント',
    warn: '気をつけて',
    danger: '大事',
  }
  return (
    <div
      style={{
        borderLeft: `3px solid ${borderColor[tone]}`,
        background: 'var(--c-bg)',
        padding: '10px 14px',
        fontSize: '12.5px',
        color: 'var(--c-ink-dim)',
      }}
    >
      <span
        style={{
          fontWeight: 700,
          color: borderColor[tone],
          marginRight: '6px',
          fontFamily: 'var(--f-mono)',
          fontSize: '11px',
          letterSpacing: '0.1em',
          textTransform: 'uppercase',
        }}
      >
        {labels[tone]}:
      </span>
      {children}
    </div>
  )
}

export function HelpPanel() {
  return (
    <div className="space-y-3">
      <header className="px-1">
        <h2
          style={{
            fontFamily: 'var(--f-display)',
            fontSize: '20px',
            fontWeight: 400,
            color: 'var(--c-ink)',
          }}
        >
          ヘルプ
        </h2>
        <p
          style={{
            marginTop: '6px',
            fontFamily: 'var(--f-mono)',
            fontSize: '12px',
            color: 'var(--c-ink-mute)',
            lineHeight: 1.7,
          }}
        >
          このアプリの使い方と、それぞれの操作で何が起きるかを、画面の動きに沿って説明します。
          見たい項目をクリックして開いてください。
        </p>
      </header>

      <Section id="overview" title="このアプリでできること" defaultOpen>
        <p>
          YouTube ライブ配信のチャット欄を見張って、参加した人の一覧を作ったり、
          特定の言葉が入ったコメントだけを集めたり、簡単な投票を集計したりできます。
        </p>
        <p>画面の上にあるタブで切り替えながら使います。</p>
        <ul className="list-disc pl-5 space-y-1">
          <li>
            <strong>名前読み上げ</strong>: 誰がチャットに参加したかの一覧を見ます。
          </li>
          <li>
            <strong>コメント検索</strong>:
            「ありがとう」「楽しい」などの言葉を含むコメントだけを抜き出します。
          </li>
          <li>
            <strong>投票</strong>:
            「A」「B」のような選択肢を決めて、視聴者に投票してもらう用途で使います。
          </li>
          <li>
            <strong>ログ</strong>:
            アプリの裏側で起きたことを時系列で見られます。うまく動かないときに役立ちます。
          </li>
          <li>
            <strong>履歴</strong>: 過去配信の視聴者一覧・コメントを閲覧します。
          </li>
          <li>
            <strong>ヘルプ</strong>: 今あなたが見ているページです。
          </li>
        </ul>
      </Section>

      <Section id="users" title="名前読み上げタブの使い方">
        <p>配信に来てくれた人の名前を一覧で見られます。挨拶や名前読み上げに使えます。</p>
        <h4 className="font-semibold pt-1">最初にやること</h4>
        <ol className="list-decimal pl-5 space-y-1">
          <li>YouTube ライブ配信のページの URL をコピーして、入力欄に貼り付けます。</li>
          <li>「切替」ボタンを押すと、その配信の見張りが始まります。</li>
          <li>あとは自動でコメントを取りに行ってくれます。</li>
        </ol>
        <h4 className="font-semibold pt-2">ボタンの意味</h4>
        <ul className="list-disc pl-5 space-y-1">
          <li>
            <strong>切替</strong>: 別の配信に乗り換えるときに押します。
          </li>
          <li>
            <strong>今すぐ取得</strong>: 自動の待ち時間を待たず、すぐにコメントを取りに行きます。
          </li>
          <li>
            <strong>リセット</strong>: 表示中の名前一覧を空にします。
          </li>
          <li>
            <strong>更新間隔</strong>: 何秒ごとに自動でコメントを取りに行くかを選びます。60 秒 / 90
            秒 / 120 秒から選べます。
          </li>
        </ul>
      </Section>

      <Section id="comments" title="コメント検索タブの使い方">
        <p>たくさんのコメントの中から、特定の言葉が入ったものだけを集めて読めます。</p>
        <h4 className="font-semibold pt-1">使い方</h4>
        <ol className="list-decimal pl-5 space-y-1">
          <li>探したい言葉を入力欄に入れて、追加します。複数登録できます。</li>
          <li>登録した言葉のどれかが入っているコメントが、一覧で表示されます。</li>
          <li>読み終わったコメントには左側のチェックをつけて、既読の目印にできます。</li>
        </ol>
        <Note tone="info">
          検索は<strong>コメント本文だけ</strong>を対象にします。投稿した人の名前は対象外です。また
          <strong>大文字と小文字は区別されます</strong>
          。ひらがな・カタカナも別の文字として扱われるので、両方探したい場合は両方登録してください。
        </Note>
      </Section>

      <Section id="votes" title="投票タブの使い方">
        <p>
          視聴者にチャットで「A」「B」のように打ってもらって、何人がどれを選んだかを数えるタブです。
        </p>
        <h4 className="font-semibold pt-1">使い方</h4>
        <ol className="list-decimal pl-5 space-y-1">
          <li>投票の選択肢になる言葉を登録します（例: 「A」「B」「C」）。</li>
          <li>視聴者にチャットでその言葉を打ってもらいます。</li>
          <li>15 秒ごとに自動で数え直されます。「再集計」ボタンですぐに数え直すこともできます。</li>
        </ol>
        <h4 className="font-semibold pt-2">数え方のルール</h4>
        <ul className="list-disc pl-5 space-y-1">
          <li>
            <strong>1 人 1 票</strong>です。同じ人が何度書き込んでも 1 票しか入りません。
          </li>
          <li>
            その人が<strong>一番最初に書いたコメント</strong>だけが投票対象になります。
          </li>
          <li>
            登録した言葉と<strong>ぴったり一致</strong>している必要があります。
          </li>
          <li>大文字と小文字、半角と全角も区別されます。</li>
          <li>前後の空白は無視されるので、うっかりスペースが入っても大丈夫です。</li>
        </ul>
        <Note tone="info">
          投票の集計は、集めたコメントをもとに毎回その場で数え直す仕組みです。確実に結果を残したい場合は、「コピー」ボタンでスプレッドシートに貼り付けておくと安心です。
        </Note>
      </Section>

      <Section id="history" title="履歴タブの使い方">
        <p>過去 7 日分の配信 snapshot を一覧で確認できます。</p>
        <ul className="list-disc pl-5 space-y-1">
          <li>一覧の行を選ぶと、その配信の視聴者一覧と全コメント（検索可）を閲覧できます。</li>
          <li>
            閲覧は<strong>読み取り専用</strong>です。現在の配信には影響しません。
          </li>
          <li>
            snapshot は<strong>自動で 7 日後に削除</strong>されます（GCS Lifecycle
            による自動管理）。
          </li>
        </ul>
      </Section>

      <Section id="logs" title="ログタブの使い方">
        <p>
          アプリの裏側で起きていることが時系列で表示されます。「うまく取れていない気がする」「エラーが出ている気がする」というときに見てください。
        </p>
        <ul className="list-disc pl-5 space-y-1">
          <li>「切替」や「今すぐ取得」を押したときの動きや、エラーメッセージが流れます。</li>
          <li>「クリア」ボタンで表示を消せます。</li>
        </ul>
      </Section>

      <Section id="clear-conditions" title="データが消える / 残る タイミング（重要）">
        <p>
          「名前の一覧が突然消えた!」と驚かないように、どの操作で何が消えて何が残るかを整理します。
        </p>
        <h4 className="font-semibold pt-1">「切替」ボタンを押したとき</h4>
        <ul className="list-disc pl-5 space-y-1">
          <li>
            <strong>名前の一覧</strong>: 空っぽになります。新しい配信の人だけを数え直すためです。
          </li>
          <li>
            <strong>集めたコメント</strong>: 残ります（コメント検索や投票でそのまま使えます）。
          </li>
        </ul>
        <Note tone="warn">
          切替直後にコメント検索タブを開くと、
          <strong>前の配信のコメントがまだ表示されている</strong>
          ことがあります。新しい配信のコメントだけを見たい場合は、コメント検索タブの「リセット」で全部隠してから使ってください。
        </Note>

        <h4 className="font-semibold pt-2">配信が終わったとき（自動で検知されます）</h4>
        <ul className="list-disc pl-5 space-y-1">
          <li>
            画面上の<strong>名前の一覧と集めたコメント</strong>は表示から消えます。
          </li>
          <li>
            消える直前に<strong>クラウドへ自動保存</strong>
            されるので、配信が終わってもデータ自体は残ります。
          </li>
          <li>自動取得も止まり、画面が「待機中」に戻ります。</li>
        </ul>
      </Section>

      <Section id="video-switch" title="動画を切替えたとき何が起きる?">
        <p>URL を変えて「切替」ボタンを押すと、その動画に合わせてデータが自動で入れ替わります。</p>
        <ul className="list-disc pl-5 space-y-1">
          <li>
            その動画の<strong>過去の参加者・コメント</strong>がクラウドから自動で復元されます。
          </li>
          <li>
            別の動画のコメントが<strong>混ざることはありません</strong>
            。コメントは動画ごとに分けて管理されています。
          </li>
        </ul>
        <Note tone="info">
          保存データがない動画（初めて見る動画など）は、空の状態から始まります。自動復元のお知らせが出ない場合は、保存データがなかったことを意味します。
        </Note>
      </Section>

      <Section id="polling" title="自動取得とコメントの数えかた">
        <ul className="list-disc pl-5 space-y-1">
          <li>
            最短でも<strong>60 秒に 1 回</strong>です。これより短い間隔は YouTube
            側の制限で設定できません。
          </li>
          <li>
            同じコメントを 2 回取りに行っても、内部で重複チェックしているので二重には数えません。
          </li>
        </ul>
      </Section>

      <Section id="errors" title="エラーが出たとき">
        <ul className="list-disc pl-5 space-y-1">
          <li>
            <strong>一時的なエラー</strong>:
            自動で何回か再試行します。それでも駄目なときだけ画面に出ます。
          </li>
          <li>
            <strong>URL や配信が見つからない</strong>: 「切替」のときに赤いメッセージで出ます。URL
            が正しいか、配信が始まっているかを確認してください。
          </li>
          <li>
            <strong>配信終了</strong>: エラーではなく自動で「待機中」に戻ります。
          </li>
        </ul>
        <Note tone="info">詳しい状況は「ログ」タブで確認できます。</Note>
      </Section>

      <Section id="common" title="そのほか共通の操作">
        <ul className="list-disc pl-5 space-y-1">
          <li>
            <strong>最後に開いていたタブを覚えている</strong>:
            次にアプリを開いたとき、前と同じタブから始まります。
          </li>
          <li>
            <strong>エラー表示</strong>: 何か問題があると、各タブの上に赤い帯で出ます。
          </li>
        </ul>
      </Section>
    </div>
  )
}
