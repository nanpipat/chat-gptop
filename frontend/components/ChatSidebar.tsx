'use client'

import { useState } from 'react'
import type { Chat } from '@/lib/api'

interface Props {
  chats: Chat[]
  selectedChatId: string | null
  onSelectChat: (id: string | null) => void
  onNewChat: () => void
}

export default function ChatSidebar({ chats, selectedChatId, onSelectChat, onNewChat }: Props) {
  const [hoveredId, setHoveredId] = useState<string | null>(null)

  return (
    <div style={{
      width: 260,
      background: '#171717',
      color: '#e0e0e0',
      display: 'flex',
      flexDirection: 'column',
      flexShrink: 0,
      height: '100vh',
    }}>
      <div style={{ padding: '16px 12px 8px' }}>
        <button
          onClick={onNewChat}
          onMouseEnter={e => (e.currentTarget.style.background = '#2a2a2a')}
          onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
          style={{
            width: '100%',
            padding: '10px 14px',
            background: 'transparent',
            color: '#e0e0e0',
            border: '1px solid #333',
            borderRadius: 8,
            cursor: 'pointer',
            fontSize: 14,
            fontWeight: 500,
            textAlign: 'left',
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            transition: 'background 0.15s',
          }}
        >
          <span style={{ fontSize: 18, lineHeight: 1 }}>+</span>
          New Chat
        </button>
      </div>

      <div style={{
        flex: 1,
        overflowY: 'auto',
        padding: '4px 8px',
      }}>
        {chats.map(chat => {
          const isActive = chat.id === selectedChatId
          const isHovered = chat.id === hoveredId
          return (
            <div
              key={chat.id}
              onClick={() => onSelectChat(chat.id)}
              onMouseEnter={() => setHoveredId(chat.id)}
              onMouseLeave={() => setHoveredId(null)}
              style={{
                padding: '10px 12px',
                cursor: 'pointer',
                borderRadius: 8,
                marginBottom: 1,
                background: isActive ? '#2a2a2a' : isHovered ? '#1e1e1e' : 'transparent',
                fontSize: 14,
                whiteSpace: 'nowrap',
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                color: isActive ? '#fff' : '#b0b0b0',
                transition: 'background 0.15s',
              }}
            >
              {chat.title || 'New Chat'}
            </div>
          )
        })}
      </div>
    </div>
  )
}
