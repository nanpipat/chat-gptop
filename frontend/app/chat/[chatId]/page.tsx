'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'

export default function ChatPage({ params }: { params: { chatId: string } }) {
  const router = useRouter()

  useEffect(() => {
    // Redirect to main page - chat selection is handled in the main UI
    router.push('/')
  }, [router])

  return null
}
