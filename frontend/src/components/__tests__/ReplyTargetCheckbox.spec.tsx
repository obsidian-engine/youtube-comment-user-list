import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { server } from '../../mocks/setup'
import { UserTable } from '../UserTable'
import { resetReplyTargetsForTest } from '../../hooks/useReplyTargets'

const users = [{ channelId: 'UCabc', displayName: 'Alice', commentCount: 1 }]

describe('自動返信のチェックボックス', () => {
  beforeEach(() => {
    resetReplyTargetsForTest()
  })

  it('保存済みのフラグがオンで表示され、クリックで解除を保存する', async () => {
    let saved: boolean | null = null
    server.use(
      http.get('*/reply-targets', () =>
        HttpResponse.json({ targets: [{ channelId: 'UCabc', setAt: '2026-11-25T00:00:00Z' }] }),
      ),
      http.put('*/reply-targets/:channelId', async ({ request }) => {
        saved = ((await request.json()) as { enabled: boolean }).enabled
        return HttpResponse.json({ enabled: saved })
      }),
    )
    render(<UserTable users={users} />)
    const box = screen.getByRole('checkbox', { name: 'Aliceさんを自動返信の対象にする' })
    await waitFor(() => expect(box).toBeChecked())

    fireEvent.click(box)
    await waitFor(() => expect(saved).toBe(false))
    expect(box).not.toBeChecked()
  })

  it('保存に失敗したら元の状態に戻し、失敗を表示する', async () => {
    server.use(
      http.get('*/reply-targets', () =>
        HttpResponse.json({ targets: [{ channelId: 'UCabc', setAt: '2026-11-25T00:00:00Z' }] }),
      ),
      http.put('*/reply-targets/:channelId', () =>
        HttpResponse.json({ error: 'internal', httpCode: 500 }, { status: 500 }),
      ),
    )
    render(<UserTable users={users} />)
    const box = screen.getByRole('checkbox', { name: 'Aliceさんを自動返信の対象にする' })
    // 読み込みが終わってから操作する
    await waitFor(() => expect(box).toBeChecked())

    fireEvent.click(box)
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument())
    expect(box).toBeChecked()
  })
})
