'use client'

import { useState, useEffect, useRef } from 'react'
import ChatSidebar from '@/components/ChatSidebar'
import ChatWindow from '@/components/ChatWindow'
import ProjectPanel from '@/components/ProjectPanel'
import type { Chat, Message, Project } from '@/lib/api'
import { getChats, createChat, getMessages, getProjects, updateChatProjects } from '@/lib/api'

export default function Home() {
  const [chats, setChats] = useState<Chat[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [selectedChatId, setSelectedChatId] = useState<string | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [selectedProjectIds, setSelectedProjectIds] = useState<string[]>([])
  const [leftCollapsed, setLeftCollapsed] = useState(false)
  const [rightCollapsed, setRightCollapsed] = useState(true)

  // Ref to always have the latest selectedChatId (avoids stale closures)
  const selectedChatIdRef = useRef(selectedChatId)
  selectedChatIdRef.current = selectedChatId

  // Skip loadMessages when a new chat is created during send
  // (the messages are managed locally during streaming)
  const skipNextLoadRef = useRef(false)

  useEffect(() => {
    loadChats()
    loadProjects()
  }, [])

  useEffect(() => {
    if (skipNextLoadRef.current) {
      skipNextLoadRef.current = false
      return
    }
    if (selectedChatId) {
      loadMessages(selectedChatId)
      // Restore project selections from the selected chat
      const chat = chats.find(c => c.id === selectedChatId)
      if (chat && chat.project_ids) {
        setSelectedProjectIds(chat.project_ids)
      }
    } else {
      setMessages([])
    }
  }, [selectedChatId])

  async function loadChats() {
    const data = await getChats()
    setChats(data)
  }

  async function loadProjects() {
    const data = await getProjects()
    setProjects(data)
  }

  async function loadMessages(chatId: string) {
    const data = await getMessages(chatId)
    setMessages(data)
  }

  function handleNewChat() {
    setSelectedChatId(null)
    setMessages([])
    setSelectedProjectIds([])
  }

  function toggleProject(id: string) {
    setSelectedProjectIds(prev => {
      const next = prev.includes(id) ? prev.filter(p => p !== id) : [...prev, id]
      // Persist to backend if there's an active chat
      const chatId = selectedChatIdRef.current
      if (chatId) {
        updateChatProjects(chatId, next).catch(console.error)
      }
      return next
    })
  }

  return (
    <div style={{ display: 'flex', height: '100vh', background: '#212121' }}>
      <ChatSidebar
        chats={chats}
        selectedChatId={selectedChatId}
        onSelectChat={setSelectedChatId}
        onNewChat={handleNewChat}
        collapsed={leftCollapsed}
        onToggle={() => setLeftCollapsed(!leftCollapsed)}
        onChatsChanged={loadChats}
      />
      <ChatWindow
        chatId={selectedChatId}
        messages={messages}
        setMessages={setMessages}
        selectedProjectIds={selectedProjectIds}
        onToggleProject={toggleProject}
        projects={projects}
        onMessagesChanged={() => {
          loadChats()
          loadProjects()
          // Use ref to get the LATEST selectedChatId (not stale closure)
          const id = selectedChatIdRef.current
          if (id) loadMessages(id)
        }}
        onChatCreated={(newChatId) => {
          skipNextLoadRef.current = true // Don't load messages in useEffect
          setSelectedChatId(newChatId)
          loadChats()
          // Persist current project selection to the new chat
          if (selectedProjectIds.length > 0) {
            updateChatProjects(newChatId, selectedProjectIds).catch(console.error)
          }
        }}
        onToggleProjectPanel={() => setRightCollapsed(!rightCollapsed)}
      />
      <ProjectPanel
        collapsed={rightCollapsed}
        onToggle={() => setRightCollapsed(!rightCollapsed)}
        selectedProjectIds={selectedProjectIds}
        onToggleProject={toggleProject}
      />
    </div>
  )
}
