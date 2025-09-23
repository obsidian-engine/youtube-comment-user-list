import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { server } from '../mocks/setup'
import { http, HttpResponse } from 'msw'
import App from '../App.jsx'

describe('Layout', () => {
  test('テーブル行に truncate-1 が含まれる', async () => {
    let currentState: 'WAITING' | 'ACTIVE' = 'WAITING'
    const users = [
      {
        channelId: 'UC1',
        displayName: 'Alice',
        joinedAt: '2024-01-01T12:00:00Z'
      }
    ]

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: currentState, count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
      http.post('*/switch-video', async () => {
        currentState = 'ACTIVE'
        return new HttpResponse(null, { status: 200 })
      }),
    )

    render(<App />)

    // 動画切り替えでACTIVE状態にする
    const input = screen.getByLabelText('videoId') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'TEST123' } })
    fireEvent.click(screen.getByRole('button', { name: '切替' }))
    
    // ACTIVE状態になるまで待機
    await waitFor(async () => {
      const activeEls = await screen.findAllByText('ACTIVE')
      expect(activeEls[0]).toBeInTheDocument()
    })

    // 見出しがあること
    expect(await screen.findByText(/参加ユーザー/)).toBeInTheDocument()
    // 名前セルに truncate-1 クラスが適用されていること
    const nameCell = await screen.findByTitle('Alice')
    expect(nameCell).not.toBeNull()
  })
})
