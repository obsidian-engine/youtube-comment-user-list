import { render, screen, fireEvent } from '@testing-library/react'
import { describe, test, expect, vi } from 'vitest'
import { Tabs } from '../Tabs'
import type { TabType } from '../Tabs'

const TABS: TabType[] = ['users', 'comments', 'votes', 'logs', 'history', 'help']

describe('Tabs', () => {
  test('アクティブタブに aria-selected=true が付く', () => {
    render(<Tabs activeTab="users" onTabChange={vi.fn()} />)
    const activeBtn = screen.getByRole('tab', { name: '名前読み上げ' })
    expect(activeBtn).toHaveAttribute('aria-selected', 'true')
    expect(activeBtn).toHaveAttribute('tabIndex', '0')
  })

  test('非アクティブタブは tabIndex=-1', () => {
    render(<Tabs activeTab="users" onTabChange={vi.fn()} />)
    const inactiveBtn = screen.getByRole('tab', { name: 'コメント検索' })
    expect(inactiveBtn).toHaveAttribute('tabIndex', '-1')
  })

  test('ArrowRight で次のタブに onTabChange が呼ばれる', () => {
    const onChange = vi.fn()
    render(<Tabs activeTab="users" onTabChange={onChange} />)
    const firstTab = screen.getByRole('tab', { name: '名前読み上げ' })
    fireEvent.keyDown(firstTab, { key: 'ArrowRight' })
    expect(onChange).toHaveBeenCalledWith('comments')
  })

  test('ArrowLeft で前のタブに wrapping される', () => {
    const onChange = vi.fn()
    render(<Tabs activeTab="users" onTabChange={onChange} />)
    const firstTab = screen.getByRole('tab', { name: '名前読み上げ' })
    fireEvent.keyDown(firstTab, { key: 'ArrowLeft' })
    expect(onChange).toHaveBeenCalledWith(TABS[TABS.length - 1])
  })

  test('Home キーで先頭タブに移動', () => {
    const onChange = vi.fn()
    render(<Tabs activeTab="logs" onTabChange={onChange} />)
    const logsTab = screen.getByRole('tab', { name: 'ログ' })
    fireEvent.keyDown(logsTab, { key: 'Home' })
    expect(onChange).toHaveBeenCalledWith('users')
  })

  test('End キーで末尾タブに移動', () => {
    const onChange = vi.fn()
    render(<Tabs activeTab="users" onTabChange={onChange} />)
    const firstTab = screen.getByRole('tab', { name: '名前読み上げ' })
    fireEvent.keyDown(firstTab, { key: 'End' })
    expect(onChange).toHaveBeenCalledWith('help')
  })
})
