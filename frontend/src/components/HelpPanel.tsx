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
    <section className="rounded-lg ring-1 ring-black/5 dark:ring-white/10 bg-white/70 dark:bg-white/5 backdrop-blur overflow-hidden">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        aria-expanded={open}
        aria-controls={`help-${id}`}
        className="w-full flex items-center justify-between gap-3 px-4 py-3 text-left hover:bg-slate-50/80 dark:hover:bg-white/5 transition"
      >
        <span className="flex items-center gap-2 font-semibold text-slate-900 dark:text-slate-100">
          <svg
            className={`w-4 h-4 transition-transform duration-200 ${open ? 'rotate-90' : ''}`}
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
          className="px-5 pb-5 pt-1 text-[13px] leading-relaxed text-slate-700 dark:text-slate-200 space-y-3"
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
  const styles: Record<string, string> = {
    info: 'bg-sky-50 dark:bg-sky-500/10 text-sky-900 dark:text-sky-200 ring-sky-200/60 dark:ring-sky-400/30',
    warn: 'bg-amber-50 dark:bg-amber-500/10 text-amber-900 dark:text-amber-200 ring-amber-200/60 dark:ring-amber-400/30',
    danger:
      'bg-rose-50 dark:bg-rose-500/10 text-rose-900 dark:text-rose-200 ring-rose-200/60 dark:ring-rose-400/30',
  }
  const labels: Record<string, string> = {
    info: 'ポイント',
    warn: '気をつけて',
    danger: '大事',
  }
  return (
    <div className={`rounded-md ring-1 px-3 py-2 text-[12.5px] ${styles[tone]}`}>
      <span className="font-semibold mr-1">{labels[tone]}:</span>
      {children}
    </div>
  )
}

export function HelpPanel() {
  return (
    <div className="space-y-3">
      <header className="px-1">
        <h2 className="text-lg font-semibold text-slate-900 dark:text-slate-100">ヘルプ</h2>
        <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">
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
        <h4 className="font-semibold pt-2">画面に出る情報</h4>
        <ul className="list-disc pl-5 space-y-1">
          <li>名前 / 発言した回数 / 最初にコメントした時刻 / 最後にコメントした時刻。</li>
          <li>表の見出し（発言数や初回時刻）をクリックすると、その順番で並べ替えできます。</li>
          <li>スマホなどの狭い画面では、見やすいように一部の列だけが表示されます。</li>
        </ul>
        <h4 className="font-semibold pt-2">画面上部の数字カード</h4>
        <ul className="list-disc pl-5 space-y-1">
          <li>参加した人の総数や、見張りを始めてからの時間が出ます。</li>
          <li>
            「スキップ数」は、YouTube から届いたコメントの一部が情報不足で使えなかった件数です。
            ほとんどの場合 0 のままで、気にしなくて大丈夫です。
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
        <h4 className="font-semibold pt-2">ボタンと操作の意味</h4>
        <ul className="list-disc pl-5 space-y-1">
          <li>
            <strong>キーワードを追加</strong>: 探したい言葉を増やします。タグの × で消せます。
          </li>
          <li>
            <strong>検索</strong>: その場ですぐに最新のコメントを取りに行きます。
          </li>
          <li>
            <strong>更新間隔</strong>: 何秒ごとに自動で取りに行くかを選びます。
          </li>
          <li>
            <strong>リセット</strong>:
            いま表示されているコメントを「読んだ扱い」にして全部隠します。
            登録した言葉自体はそのままです。
          </li>
        </ul>
        <Note tone="info">
          検索は<strong>コメント本文だけ</strong>を対象にします。投稿した人の名前は対象外です。 また
          <strong>大文字と小文字は区別されます</strong>
          。たとえば「abc」で検索しても「ABC」は出ません。
          ひらがな・カタカナも別の文字として扱われるので、両方探したい場合は両方登録してください。
        </Note>
        <Note tone="info">
          「読んだ扱い」にしたコメントは、お使いの端末・ブラウザにだけ覚えられます。
          別のパソコンやスマホで開くと、また表示されます。
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
            あとから気が変わって書き直しても反映されません。
          </li>
          <li>
            登録した言葉と<strong>ぴったり一致</strong>している必要があります。
            「A」が選択肢なら「A」と打った人だけが対象で、「A です」「A!」は数えません。
          </li>
          <li>大文字と小文字、半角と全角も区別されます。</li>
          <li>前後の空白は無視されるので、うっかりスペースが入っても大丈夫です。</li>
        </ul>
        <h4 className="font-semibold pt-2">投票した人を確認する / コピーする</h4>
        <ul className="list-disc pl-5 space-y-1">
          <li>
            集計結果の<strong>選択肢の行をクリック</strong>
            すると、その選択肢に投票した人の一覧が開きます。 表示名とチャンネル ID が並びます。
          </li>
          <li>
            一覧の<strong>「コピー」ボタン</strong>を押すと、表示名 / チャンネル ID をタブ区切りで
            クリップボードにコピーできます。Google スプレッドシートや Excel
            にそのまま貼り付けて表になります。
          </li>
        </ul>
        <h4 className="font-semibold pt-2">登録した選択肢の保存</h4>
        <ul className="list-disc pl-5 space-y-1">
          <li>
            登録した投票の選択肢は、お使いの端末・ブラウザに自動で保存されます。
            タブを閉じても、ブラウザを開き直しても、選択肢はそのまま残ります。
          </li>
          <li>票数や投票者の一覧は保存対象外で、開き直したときは自動で数え直します。</li>
        </ul>
        <Note tone="info">
          投票の集計は、集めたコメントをもとに毎回その場で数え直す仕組みです。
          コメントはサーバーに保存されているので、再起動後でも再集計すれば結果を再現できます。
          確実に結果を残したい場合は、「コピー」ボタンでスプレッドシートに貼り付けておくと安心です。
        </Note>
      </Section>

      <Section id="logs" title="ログタブの使い方">
        <p>
          アプリの裏側で起きていることが時系列で表示されます。
          「うまく取れていない気がする」「エラーが出ている気がする」というときに見てください。
        </p>
        <ul className="list-disc pl-5 space-y-1">
          <li>「切替」や「今すぐ取得」を押したときの動きや、エラーメッセージが流れます。</li>
          <li>「クリア」ボタンで表示を消せます。</li>
        </ul>
        <Note tone="info">
          いまは「コメントを取りに行ったとき」のログが中心です。それ以外の操作のログは順次対応予定です。
        </Note>
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
          <strong>前の配信のコメントがまだ表示されている</strong>ことがあります。
          新しい配信のコメントだけを見たい場合は、コメント検索タブの「リセット」で全部隠してから使ってください。
        </Note>

        <h4 className="font-semibold pt-2">「リセット」ボタン（名前読み上げタブ）</h4>
        <ul className="list-disc pl-5 space-y-1">
          <li>
            <strong>名前の一覧</strong>: 空っぽになります。
          </li>
          <li>
            <strong>集めたコメント</strong>: 残ります。
          </li>
          <li>見張り状態が「待機中」に戻り、自動取得も止まります。</li>
        </ul>

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
        <Note tone="info">
          サーバーが再起動した場合は、保存されたデータから自動で復元されます。
          ただし画面に再表示するための機能は今のところ用意されていないため、
          「いま画面で見ているデータ」を確認したい場合はスクリーンショットなどで手元に残すと安心です。
        </Note>

        <h4 className="font-semibold pt-2">アプリのサーバーが再起動したとき</h4>
        <ul className="list-disc pl-5 space-y-1">
          <li>クラウドに保存する仕組みが有効になっていれば、直前の状態に戻ります。</li>
          <li>有効になっていない場合は、データはすべて消えた状態で立ち上がります。</li>
          <li>
            復元に成功すると、画面の上に
            <strong>「○時○分時点の保存データを復元しました」というお知らせ</strong>
            が一度だけ表示されます。これでいつ時点のデータが戻ったかが分かります。
          </li>
        </ul>
        <Note tone="info">
          保存は「切替」「リセット」を押したタイミングと、配信中はおよそ 60
          秒ごとに自動で行われます。 そのため、急に強制終了したような場合はごく直近の最大 60
          秒分が失われる可能性があります。
        </Note>

        <h4 className="font-semibold pt-2">同じ人が名前を途中で変えた場合</h4>
        <ul className="list-disc pl-5 space-y-1">
          <li>
            名前の欄は、その人の<strong>最新の名前</strong>で表示されます。
          </li>
          <li>「最初にコメントした時刻」は変わりません。</li>
        </ul>
      </Section>

      <Section id="video-switch" title="🔄 動画を切替えたとき何が起きる?">
        <p>URL を変えて「切替」ボタンを押すと、その動画に合わせてデータが自動で入れ替わります。</p>
        <h4 className="font-semibold pt-1">切替えたときの動き</h4>
        <ul className="list-disc pl-5 space-y-1">
          <li>
            その動画の<strong>過去の参加者・コメント</strong>
            がクラウドから自動で復元されます。以前に同じ動画を見ていたときの続きから再開できます。
          </li>
          <li>
            復元が完了すると、画面の上に
            <strong>「○時○分時点の保存データを復元しました」というお知らせ</strong>
            が一度だけ表示されます。いつ時点のデータが戻ったかが分かります。
          </li>
          <li>
            別の動画のコメントが<strong>混ざることはありません</strong>
            。コメントは動画ごとに分けて管理されています。
          </li>
        </ul>
        <h4 className="font-semibold pt-2">「動画 A → B → A」と行き来しても大丈夫?</h4>
        <ul className="list-disc pl-5 space-y-1">
          <li>
            動画 A に戻したとき、A のデータは<strong>そのまま残っています</strong>。 動画 B
            を見ていた間も、A のデータは上書きされません。
          </li>
          <li>
            配信が終わった動画でも、
            <strong>7 日以内であれば</strong>切替えると以前の状態を復元できます。
          </li>
        </ul>
        <Note tone="info">
          保存データがない動画（初めて見る動画など）は、空の状態から始まります。
          自動復元のお知らせが出ない場合は、保存データがなかったことを意味します。
        </Note>
      </Section>

      <Section id="polling" title="自動取得とコメントの数えかた">
        <h4 className="font-semibold">どれくらいの間隔で取りに行くの?</h4>
        <ul className="list-disc pl-5 space-y-1">
          <li>
            最短でも<strong>60 秒に 1 回</strong>です。これより短い間隔は YouTube
            側の制限で設定できません。
          </li>
          <li>
            選べる間隔（60 秒 / 90 秒 / 120 秒）はこの制限に合わせてあります。 長い配信では 90 秒や
            120 秒にすると、サーバーへの負担を減らせます。
          </li>
        </ul>
        <h4 className="font-semibold pt-2">同じコメントが二重に数えられない?</h4>
        <ul className="list-disc pl-5 space-y-1">
          <li>
            同じコメントを 2 回取りに行っても、内部で重複チェックしているので二重には数えません。
          </li>
        </ul>
        <h4 className="font-semibold pt-2">削除されたコメントはどうなる?</h4>
        <ul className="list-disc pl-5 space-y-1">
          <li>
            投稿者や配信者が削除したコメントは、そもそも YouTube
            から届かないため、このアプリにも残りません。
          </li>
        </ul>
        <h4 className="font-semibold pt-2">配信を始めていないときに「今すぐ取得」を押したら?</h4>
        <ul className="list-disc pl-5 space-y-1">
          <li>何もしないで終わります。無駄なアクセスはしないので安心して押せます。</li>
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
            <strong>配信終了</strong>:
            エラーではなく自動で「待機中」に戻ります（前項の通り、データは消えます）。
          </li>
        </ul>
        <Note tone="info">詳しい状況は「ログ」タブで確認できます。</Note>
      </Section>

      <Section id="common" title="そのほか共通の操作">
        <ul className="list-disc pl-5 space-y-1">
          <li>
            <strong>テーマ切替</strong>: 右上のボタンで明るい /
            暗いを切り替えられます。次回も覚えています。
          </li>
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
