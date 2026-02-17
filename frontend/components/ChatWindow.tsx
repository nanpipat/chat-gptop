'use client'

import { useState, useRef, useEffect } from 'react'
import MessageBubble from './MessageBubble'
import ProjectPanel from './ProjectPanel'
import type { Message, Project } from '@/lib/api'
import { sendMessage, createChat } from '@/lib/api'

interface Props {
  chatId: string | null
  messages: Message[]
  setMessages: React.Dispatch<React.SetStateAction<Message[]>>
  selectedProjectIds: string[]
  onToggleProject: (id: string) => void
  projects: Project[]
  onMessagesChanged: () => void
  onChatCreated: (chatId: string) => void
}

export default function ChatWindow({
  chatId,
  messages,
  setMessages,
  selectedProjectIds,
  onToggleProject,
  projects,
  onMessagesChanged,
  onChatCreated,
}: Props) {
  const [input, setInput] = useState('')
  const [streaming, setStreaming] = useState(false)
  const [streamingContent, setStreamingContent] = useState('')
  const [showProjectPanel, setShowProjectPanel] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, streamingContent])

  const isLanding = !chatId && messages.length === 0 && !streaming

  async function handleSend() {
    if (!input.trim() || streaming) return

    let currentChatId = chatId
    if (!currentChatId) {
      const chat = await createChat(input.trim().slice(0, 50))
      currentChatId = chat.id
      onChatCreated(currentChatId)
    }

    const userMsg: Message = {
      id: 'temp-' + Date.now(),
      chat_id: currentChatId,
      role: 'user',
      content: input.trim(),
      created_at: new Date().toISOString(),
    }

    setMessages(prev => [...prev, userMsg])
    setInput('')
    setStreaming(true)
    setStreamingContent('')

    sendMessage(
      currentChatId,
      userMsg.content,
      selectedProjectIds,
      (token) => {
        setStreamingContent(prev => prev + token)
      },
      () => {
        setStreaming(false)
        setStreamingContent('')
        onMessagesChanged()
      },
      (err) => {
        console.error('Stream error:', err)
        setStreaming(false)
        setStreamingContent('')
      },
    )
  }

  const selectedProjects = projects.filter(p => selectedProjectIds.includes(p.id))

  const inputArea = (
    <div style={{
      maxWidth: 768,
      margin: '0 auto',
      width: '100%',
    }}>
      {/* Selected project pills */}
      {selectedProjects.length > 0 && (
        <div style={{
          display: 'flex',
          gap: 6,
          flexWrap: 'wrap',
          marginBottom: 8,
          paddingLeft: 4,
        }}>
          {selectedProjects.map(p => (
            <span
              key={p.id}
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 4,
                padding: '3px 10px',
                background: '#2a2a2a',
                border: '1px solid #333',
                borderRadius: 12,
                fontSize: 12,
                color: '#d4a574',
              }}
            >
              {p.name}
              <button
                onClick={() => onToggleProject(p.id)}
                style={{
                  background: 'none',
                  border: 'none',
                  color: '#888',
                  cursor: 'pointer',
                  fontSize: 14,
                  padding: 0,
                  lineHeight: 1,
                  marginLeft: 2,
                }}
              >
                ×
              </button>
            </span>
          ))}
        </div>
      )}

      {/* Input box */}
      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: 8,
        background: '#2a2a2a',
        borderRadius: 24,
        padding: '10px 8px 10px 16px',
        border: '1px solid #333',
        transition: 'border-color 0.2s',
      }}>
        {/* Project button */}
        <button
          onClick={() => setShowProjectPanel(true)}
          title="Manage projects"
          style={{
            background: 'none',
            border: 'none',
            color: selectedProjectIds.length > 0 ? '#d4a574' : '#666',
            cursor: 'pointer',
            fontSize: 18,
            padding: '4px 6px',
            borderRadius: 6,
            lineHeight: 1,
            flexShrink: 0,
            transition: 'color 0.15s',
          }}
          onMouseEnter={e => (e.currentTarget.style.color = '#d4a574')}
          onMouseLeave={e => (e.currentTarget.style.color = selectedProjectIds.length > 0 ? '#d4a574' : '#666')}
        >
          &#128194;
        </button>

        <input
          ref={inputRef}
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={e => {
            if (e.key === 'Enter' && !e.shiftKey) {
              e.preventDefault()
              handleSend()
            }
          }}
          placeholder="Message..."
          disabled={streaming}
          style={{
            flex: 1,
            background: 'transparent',
            border: 'none',
            outline: 'none',
            color: '#e0e0e0',
            fontSize: 15,
          }}
        />

        {/* Send button */}
        <button
          onClick={handleSend}
          disabled={streaming || !input.trim()}
          style={{
            background: streaming || !input.trim() ? '#333' : '#d4a574',
            color: streaming || !input.trim() ? '#666' : '#1a1a1a',
            border: 'none',
            borderRadius: '50%',
            width: 36,
            height: 36,
            cursor: streaming || !input.trim() ? 'not-allowed' : 'pointer',
            fontSize: 16,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexShrink: 0,
            transition: 'background 0.15s',
          }}
        >
          &#8593;
        </button>
      </div>
    </div>
  )

  return (
    <div style={{
      flex: 1,
      display: 'flex',
      flexDirection: 'column',
      background: '#212121',
      height: '100vh',
      overflow: 'hidden',
    }}>
      {isLanding ? (
        /* Landing state */
        <div style={{
          flex: 1,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '0 24px',
        }}>
          <h1 style={{
            fontSize: 32,
            fontWeight: 600,
            color: '#e0e0e0',
            marginBottom: 48,
            textAlign: 'center',
          }}>
            What can I help you with?
          </h1>
          <div style={{ width: '100%', maxWidth: 768 }}>
            {inputArea}
          </div>
        </div>
      ) : (
        /* Active chat state */
        <>
          <div style={{
            flex: 1,
            overflowY: 'auto',
            paddingTop: 16,
            paddingBottom: 8,
          }}>
            <div style={{ maxWidth: 768, margin: '0 auto' }}>
              {messages.map(msg => (
                <MessageBubble key={msg.id} role={msg.role} content={msg.content} />
              ))}
              {streaming && streamingContent && (
                <MessageBubble role="assistant" content={streamingContent} />
              )}
              {streaming && !streamingContent && (
                <div style={{
                  padding: '16px 24px',
                  maxWidth: 768,
                  margin: '0 auto',
                  display: 'flex',
                  gap: 14,
                }}>
                  <div style={{
                    width: 28,
                    height: 28,
                    borderRadius: '50%',
                    background: 'linear-gradient(135deg, #d4a574, #c49060)',
                    flexShrink: 0,
                  }} />
                  <div style={{ color: '#888', fontSize: 15 }}>Thinking...</div>
                </div>
              )}
              <div ref={bottomRef} />
            </div>
          </div>
          <div style={{
            padding: '12px 24px 24px',
            background: '#212121',
          }}>
            {inputArea}
          </div>
        </>
      )}

      {/* Project panel slide-over */}
      <ProjectPanel
        open={showProjectPanel}
        onClose={() => setShowProjectPanel(false)}
        selectedProjectIds={selectedProjectIds}
        onToggleProject={onToggleProject}
      />
    </div>
  )
}
