import { render, screen, fireEvent } from '@testing-library/react'
import { vi } from 'vitest'
import { server } from '../mocks/setup'
import { http, HttpResponse, delay } from 'msw'
import App from '../App.jsx'

// Note: matchMedia and localStorage are mocked globally in setup.ts

describe('Pending states', () => {
  test('切替ボタンが pending 中は disabled/aria-busy', async () => {
    server.use(
      http.post('*/switch-video', async () => {
        await delay(200)
        return new HttpResponse(null, { status: 200 })
      }),
      http.get('*/status', () => HttpResponse.json({ status: 'ACTIVE', count: 0 })),
      http.get('*/users.json', () => HttpResponse.json([])),
    )

    render(<App />)

    const input = await screen.findByLabelText('videoId')
    fireEvent.change(input, { target: { value: 'VID' } })

    const btn = screen.getByRole('button', { name: '切替' })
    fireEvent.click(btn)
    expect(btn).toBeDisabled()
    expect(btn).toHaveAttribute('aria-busy', 'true')
  })
})

