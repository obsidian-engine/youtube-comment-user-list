import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { ThemeToggle } from '../ThemeToggle'

// matchMediaのモック
const mockMatchMedia = vi.fn().mockImplementation(query => ({
  matches: false,
  media: query,
  onchange: null,
  addListener: vi.fn(),
  removeListener: vi.fn(),
  addEventListener: vi.fn(),
  removeEventListener: vi.fn(),
  dispatchEvent: vi.fn(),
}))

// localStorageのモック
const mockLocalStorage = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
}

describe('ThemeToggle', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    
    // matchMediaをグローバルに設定
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: mockMatchMedia,
    })
    
    // localStorageをグローバルに設定
    Object.defineProperty(window, 'localStorage', {
      writable: true,
      value: mockLocalStorage,
    })
    
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
    
    const button = screen.getByRole('button', { name: /テーマを切り替え/ })
    expect(button).toBeInTheDocument()
  })

  it('shows sun icon for light mode', () => {
    mockLocalStorage.getItem.mockReturnValue('light')
    mockMatchMedia.mockReturnValue({ matches: false })
    
    render(<ThemeToggle />)
    
    const sunIcon = screen.getByLabelText('Light mode')
    expect(sunIcon).toBeInTheDocument()
  })

  it('shows moon icon for dark mode', () => {
    mockLocalStorage.getItem.mockReturnValue('dark')
    mockMatchMedia.mockReturnValue({ matches: false })
    
    render(<ThemeToggle />)
    
    const moonIcon = screen.getByLabelText('Dark mode')
    expect(moonIcon).toBeInTheDocument()
  })

  it('toggles from light to dark mode when clicked', () => {
    mockLocalStorage.getItem.mockReturnValue('light')
    mockMatchMedia.mockReturnValue({ matches: false })
    
    render(<ThemeToggle />)
    
    const button = screen.getByRole('button', { name: /テーマを切り替え/ })
    fireEvent.click(button)
    
    expect(document.documentElement.classList.add).toHaveBeenCalledWith('dark')
    expect(mockLocalStorage.setItem).toHaveBeenCalledWith('theme', 'dark')
  })

  it('toggles from dark to light mode when clicked', () => {
    mockLocalStorage.getItem.mockReturnValue('dark')
    mockMatchMedia.mockReturnValue({ matches: false })
    
    render(<ThemeToggle />)
    
    const button = screen.getByRole('button', { name: /テーマを切り替え/ })
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