import { render, screen, fireEvent } from '@testing-library/react'
import { vi } from 'vitest'
import App from '../App.jsx'

// Note: matchMedia and localStorage are mocked globally in setup.ts

describe('App error banner', () => {
  test('videoId 未入力で切替するとエラーバナー表示', async () => {
    render(<App />)
    fireEvent.click(screen.getByRole('button', { name: '切替' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('videoId を入力してください。')
  })
})

