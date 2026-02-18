'use client'

import { useState, useRef, useEffect } from 'react'
import type { Chat } from '@/lib/api'
import { deleteChat } from '@/lib/api'

interface Props {
  chats: Chat[]
  selectedChatId: string | null
  onSelectChat: (id: string | null) => void
  onNewChat: () => void
  collapsed: boolean
  onToggle: () => void
  onChatsChanged: () => void
}

export default function ChatSidebar({
  chats,
  selectedChatId,
  onSelectChat,
  onNewChat,
  collapsed,
  onToggle,
  onChatsChanged,
}: Props) {
  const [hoveredId, setHoveredId] = useState<string | null>(null)
  const [menuOpenId, setMenuOpenId] = useState<string | null>(null)
  const menuRef = useRef<HTMLDivElement>(null)

  // Close menu when clicking outside
  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpenId(null)
      }
    }
    if (menuOpenId) {
      document.addEventListener('mousedown', handleClickOutside)
      return () => document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [menuOpenId])

  async function handleDelete(chatId: string) {
    setMenuOpenId(null)
    await deleteChat(chatId)
    if (selectedChatId === chatId) {
      onSelectChat(null)
    }
    onChatsChanged()
  }

  return (
    <div
      className="sidebar-transition"
      style={{
        width: collapsed ? 48 : 260,
        minWidth: collapsed ? 48 : 260,
        background: '#171717',
        color: '#e0e0e0',
        display: 'flex',
        flexDirection: 'column',
        height: '100vh',
        borderRight: '1px solid #2a2a2a',
        position: 'relative',
      }}
    >
      {/* Toggle button */}
      <div style={{
        padding: collapsed ? '12px 0' : '12px 12px 8px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: collapsed ? 'center' : 'space-between',
        gap: 8,
      }}>
        <button
          onClick={onToggle}
          title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          style={{
            background: 'none',
            border: 'none',
            color: '#888',
            cursor: 'pointer',
            fontSize: 18,
            padding: '6px 8px',
            borderRadius: 6,
            lineHeight: 1,
            flexShrink: 0,
            transition: 'color 0.15s, background 0.15s',
          }}
          onMouseEnter={e => {
            e.currentTarget.style.color = '#fff'
            e.currentTarget.style.background = '#2a2a2a'
          }}
          onMouseLeave={e => {
            e.currentTarget.style.color = '#888'
            e.currentTarget.style.background = 'none'
          }}
        >
          {collapsed ? '☰' : '◀'}
        </button>

        {!collapsed && (
          <button
            onClick={onNewChat}
            title="New Chat"
            style={{
              background: 'none',
              border: 'none',
              color: '#888',
              cursor: 'pointer',
              fontSize: 18,
              padding: '6px 8px',
              borderRadius: 6,
              lineHeight: 1,
              transition: 'color 0.15s, background 0.15s',
            }}
            onMouseEnter={e => {
              e.currentTarget.style.color = '#fff'
              e.currentTarget.style.background = '#2a2a2a'
            }}
            onMouseLeave={e => {
              e.currentTarget.style.color = '#888'
              e.currentTarget.style.background = 'none'
            }}
          >
            +
          </button>
        )}
      </div>

      {/* Chat list - hidden when collapsed */}
      <div
        className={`sidebar-content ${collapsed ? 'collapsed' : ''}`}
        style={{
          flex: 1,
          overflowY: 'auto',
          padding: collapsed ? 0 : '4px 8px',
        }}
      >
        {!collapsed && chats.map(chat => {
          const isActive = chat.id === selectedChatId
          const isHovered = chat.id === hoveredId
          return (
            <div
              key={chat.id}
              onMouseEnter={() => setHoveredId(chat.id)}
              onMouseLeave={() => setHoveredId(null)}
              style={{
                position: 'relative',
                display: 'flex',
                alignItems: 'center',
                padding: '10px 12px',
                cursor: 'pointer',
                borderRadius: 8,
                marginBottom: 1,
                background: isActive ? '#2a2a2a' : isHovered ? '#1e1e1e' : 'transparent',
                fontSize: 14,
                color: isActive ? '#fff' : '#b0b0b0',
                transition: 'background 0.15s',
              }}
            >
              <span
                onClick={() => onSelectChat(chat.id)}
                style={{
                  flex: 1,
                  whiteSpace: 'nowrap',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                }}
              >
                {chat.title || 'New Chat'}
              </span>

              {/* "..." menu button - visible on hover or when menu is open */}
              {(isHovered || menuOpenId === chat.id) && (
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    setMenuOpenId(menuOpenId === chat.id ? null : chat.id)
                  }}
                  style={{
                    background: 'none',
                    border: 'none',
                    color: '#888',
                    cursor: 'pointer',
                    fontSize: 16,
                    padding: '0 4px',
                    lineHeight: 1,
                    flexShrink: 0,
                    borderRadius: 4,
                    transition: 'color 0.15s',
                  }}
                  onMouseEnter={e => (e.currentTarget.style.color = '#fff')}
                  onMouseLeave={e => (e.currentTarget.style.color = '#888')}
                >
                  ⋯
                </button>
              )}

              {/* Dropdown menu */}
              {menuOpenId === chat.id && (
                <div
                  ref={menuRef}
                  style={{
                    position: 'absolute',
                    right: 8,
                    top: '100%',
                    zIndex: 50,
                    background: '#2a2a2a',
                    border: '1px solid #3a3a3a',
                    borderRadius: 8,
                    padding: 4,
                    boxShadow: '0 4px 16px rgba(0,0,0,0.4)',
                    minWidth: 140,
                  }}
                >
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      handleDelete(chat.id)
                    }}
                    style={{
                      width: '100%',
                      padding: '8px 12px',
                      background: 'transparent',
                      border: 'none',
                      borderRadius: 6,
                      color: '#ef4444',
                      cursor: 'pointer',
                      fontSize: 13,
                      textAlign: 'left',
                      display: 'flex',
                      alignItems: 'center',
                      gap: 8,
                      transition: 'background 0.15s',
                    }}
                    onMouseEnter={e => (e.currentTarget.style.background = '#333')}
                    onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
                  >
                    🗑 Delete
                  </button>
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
