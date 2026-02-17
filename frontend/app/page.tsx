'use client'

import { useState, useEffect } from 'react'
import ChatSidebar from '@/components/ChatSidebar'
import ChatWindow from '@/components/ChatWindow'
import type { Chat, Message, Project } from '@/lib/api'
import { getChats, createChat, getMessages, getProjects } from '@/lib/api'

export default function Home() {
  const [chats, setChats] = useState<Chat[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [selectedChatId, setSelectedChatId] = useState<string | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [selectedProjectIds, setSelectedProjectIds] = useState<string[]>([])

  useEffect(() => {
    loadChats()
    loadProjects()
  }, [])

  useEffect(() => {
    if (selectedChatId) {
      loadMessages(selectedChatId)
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
  }

  function toggleProject(id: string) {
    setSelectedProjectIds(prev =>
      prev.includes(id) ? prev.filter(p => p !== id) : [...prev, id]
    )
  }

  return (
    <div style={{ display: 'flex', height: '100vh', background: '#212121' }}>
      <ChatSidebar
        chats={chats}
        selectedChatId={selectedChatId}
        onSelectChat={setSelectedChatId}
        onNewChat={handleNewChat}
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
          if (selectedChatId) loadMessages(selectedChatId)
        }}
        onChatCreated={(newChatId) => {
          setSelectedChatId(newChatId)
          loadChats()
        }}
      />
    </div>
  )
}
