import { render, screen } from '@testing-library/react'
import { server } from '../mocks/setup'
import { http, HttpResponse } from 'msw'
import App from '../App.jsx'

describe('Layout', () => {
  test('テーブル行に truncate-1 が含まれる', async () => {
    server.use(
      http.get('*/status', () => HttpResponse.json({ status: 'ACTIVE', count: 1 })),
      http.get('*/users.json', () => HttpResponse.json(['Alice'])),
    )
    render(<App />)
    // 見出しがあること
    expect(await screen.findByText(/参加ユーザー/)).toBeInTheDocument()
    // 名前セルに truncate-1 クラスが適用されていること
    const nameCell = await screen.findByTitle('Alice')
    expect(nameCell).not.toBeNull()
  })
})
