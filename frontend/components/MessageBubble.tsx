'use client'

interface Props {
  role: string
  content: string
}

export default function MessageBubble({ role, content }: Props) {
  const isUser = role === 'user'

  if (isUser) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'flex-end',
        padding: '8px 24px',
      }}>
        <div style={{
          maxWidth: '70%',
          background: '#303030',
          color: '#e8e8e8',
          padding: '12px 16px',
          borderRadius: '18px 18px 4px 18px',
          fontSize: 15,
          lineHeight: 1.6,
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-word',
        }}>
          {content}
        </div>
      </div>
    )
  }

  return (
    <div style={{
      padding: '16px 24px',
    }}>
      <div style={{
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
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: 13,
          color: '#fff',
          fontWeight: 600,
          flexShrink: 0,
          marginTop: 2,
        }}>
          AI
        </div>
        <div style={{
          color: '#e0e0e0',
          lineHeight: 1.7,
          fontSize: 15,
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-word',
          flex: 1,
        }}>
          {content}
        </div>
      </div>
    </div>
  )
}
