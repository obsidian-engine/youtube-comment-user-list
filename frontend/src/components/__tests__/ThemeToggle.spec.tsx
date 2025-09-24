import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { ThemeToggle } from '../ThemeToggle'

// テスト用のモック変数
const mockMatchMedia = window.matchMedia as ReturnType<typeof vi.fn>
const mockLocalStorage = window.localStorage as {
  getItem: ReturnType<typeof vi.fn>
  setItem: ReturnType<typeof vi.fn>
  removeItem: ReturnType<typeof vi.fn>
  clear: ReturnType<typeof vi.fn>
}

describe('ThemeToggle', () => {
  beforeEach(() => {
    vi.clearAllMocks()

    // document.documentElementのクラスリストをモック
    document.documentElement.classList.add = vi.fn()
    document.documentElement.classList.remove = vi.fn()
  })

  afterEach(() => {
    document.documentElement.classList.add = vi.fn()
    document.documentElement.classList.remove = vi.fn()
  })

  it('renders theme toggle button', () => {
    mockLocalStorage.getItem.mockReturnValue(null)
    mockMatchMedia.mockReturnValue({ matches: false })

    render(<ThemeToggle />)

    const button = screen.getByRole('button', { name: 'ダークモードに切り替え' })
    expect(button).toBeInTheDocument()
  })

  it('shows moon icon for light mode (ダークモードに切り替え)', () => {
    mockLocalStorage.getItem.mockReturnValue('light')
    mockMatchMedia.mockReturnValue({ matches: false })

    render(<ThemeToggle />)

    const button = screen.getByRole('button', { name: 'ダークモードに切り替え' })
    expect(button).toBeInTheDocument()
  })

  it('shows sun icon for dark mode (ライトモードに切り替え)', () => {
    mockLocalStorage.getItem.mockReturnValue('dark')
    mockMatchMedia.mockReturnValue({ matches: false })

    render(<ThemeToggle />)

    const button = screen.getByRole('button', { name: 'ライトモードに切り替え' })
    expect(button).toBeInTheDocument()
  })

  it('toggles from light to dark mode when clicked', () => {
    mockLocalStorage.getItem.mockReturnValue('light')
    mockMatchMedia.mockReturnValue({ matches: false })

    render(<ThemeToggle />)

    const button = screen.getByRole('button', { name: 'ダークモードに切り替え' })
    fireEvent.click(button)

    expect(document.documentElement.classList.add).toHaveBeenCalledWith('dark')
    expect(mockLocalStorage.setItem).toHaveBeenCalledWith('theme', 'dark')
  })

  it('toggles from dark to light mode when clicked', () => {
    mockLocalStorage.getItem.mockReturnValue('dark')
    mockMatchMedia.mockReturnValue({ matches: false })

    render(<ThemeToggle />)

    const button = screen.getByRole('button', { name: 'ライトモードに切り替え' })
    fireEvent.click(button)

    expect(document.documentElement.classList.remove).toHaveBeenCalledWith('dark')
    expect(mockLocalStorage.setItem).toHaveBeenCalledWith('theme', 'light')
  })

  it('uses system preference when no saved theme', () => {
    mockLocalStorage.getItem.mockReturnValue(null)
    mockMatchMedia.mockReturnValue({ matches: true }) // システムがダークモードを優先
    
    render(<ThemeToggle />)
    
    expect(document.documentElement.classList.add).toHaveBeenCalledWith('dark')
  })

  it('uses light mode when no saved theme and system prefers light', () => {
    mockLocalStorage.getItem.mockReturnValue(null)
    mockMatchMedia.mockReturnValue({ matches: false }) // システムがライトモードを優先
    
    render(<ThemeToggle />)
    
    expect(document.documentElement.classList.remove).toHaveBeenCalledWith('dark')
  })
})