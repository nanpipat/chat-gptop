'use client'

import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark } from 'react-syntax-highlighter/dist/cjs/styles/prism'

interface Props {
  role: string
  content: string
}

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // fallback
      const ta = document.createElement('textarea')
      ta.value = text
      document.body.appendChild(ta)
      ta.select()
      document.execCommand('copy')
      document.body.removeChild(ta)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  return (
    <button
      onClick={handleCopy}
      style={{
        background: 'none',
        border: '1px solid #444',
        borderRadius: 6,
        color: copied ? '#6ee06e' : '#aaa',
        cursor: 'pointer',
        fontSize: 12,
        padding: '4px 10px',
        transition: 'all 0.15s',
        display: 'flex',
        alignItems: 'center',
        gap: 4,
      }}
      onMouseEnter={e => {
        if (!copied) e.currentTarget.style.color = '#ddd'
      }}
      onMouseLeave={e => {
        if (!copied) e.currentTarget.style.color = '#aaa'
      }}
    >
      {copied ? '✓ Copied' : '⎘ Copy'}
    </button>
  )
}

function MarkdownContent({ content }: { content: string }) {
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        code({ node, className, children, ...props }) {
          const match = /language-(\w+)/.exec(className || '')
          const codeString = String(children).replace(/\n$/, '')

          if (match) {
            return (
              <div style={{
                borderRadius: 10,
                overflow: 'hidden',
                margin: '12px 0',
                border: '1px solid #333',
                background: '#1e1e1e',
              }}>
                <div style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  padding: '6px 12px',
                  background: '#2a2a2a',
                  borderBottom: '1px solid #333',
                }}>
                  <span style={{
                    fontSize: 12,
                    color: '#888',
                    fontFamily: 'monospace',
                    textTransform: 'lowercase',
                  }}>
                    {match[1]}
                  </span>
                  <CopyButton text={codeString} />
                </div>
                <SyntaxHighlighter
                  style={oneDark as any}
                  language={match[1]}
                  PreTag="div"
                  customStyle={{
                    margin: 0,
                    padding: '14px 16px',
                    fontSize: 13,
                    lineHeight: 1.6,
                    background: '#1e1e1e',
                  }}
                >
                  {codeString}
                </SyntaxHighlighter>
              </div>
            )
          }

          // Inline code
          return (
            <code
              style={{
                background: '#2a2a2a',
                border: '1px solid #3a3a3a',
                borderRadius: 5,
                padding: '2px 6px',
                fontSize: '0.88em',
                fontFamily: '"SF Mono", "Fira Code", Menlo, monospace',
                color: '#e8b88a',
              }}
              {...props}
            >
              {children}
            </code>
          )
        },
        // Headings
        h1: ({ children }) => (
          <h1 style={{ fontSize: 22, fontWeight: 700, margin: '20px 0 10px', color: '#f0f0f0', borderBottom: '1px solid #333', paddingBottom: 6 }}>{children}</h1>
        ),
        h2: ({ children }) => (
          <h2 style={{ fontSize: 19, fontWeight: 600, margin: '18px 0 8px', color: '#f0f0f0' }}>{children}</h2>
        ),
        h3: ({ children }) => (
          <h3 style={{ fontSize: 16, fontWeight: 600, margin: '14px 0 6px', color: '#e8e8e8' }}>{children}</h3>
        ),
        // Paragraphs
        p: ({ children }) => (
          <p style={{ margin: '8px 0', lineHeight: 1.7 }}>{children}</p>
        ),
        // Lists
        ul: ({ children }) => (
          <ul style={{ margin: '8px 0', paddingLeft: 22 }}>{children}</ul>
        ),
        ol: ({ children }) => (
          <ol style={{ margin: '8px 0', paddingLeft: 22 }}>{children}</ol>
        ),
        li: ({ children }) => (
          <li style={{ margin: '4px 0', lineHeight: 1.6 }}>{children}</li>
        ),
        // Blockquote
        blockquote: ({ children }) => (
          <blockquote style={{
            borderLeft: '3px solid #d4a574',
            margin: '12px 0',
            padding: '4px 16px',
            color: '#bbb',
            background: '#252525',
            borderRadius: '0 6px 6px 0',
          }}>
            {children}
          </blockquote>
        ),
        // Tables
        table: ({ children }) => (
          <div style={{ overflowX: 'auto', margin: '12px 0' }}>
            <table style={{
              borderCollapse: 'collapse',
              width: '100%',
              fontSize: 14,
            }}>
              {children}
            </table>
          </div>
        ),
        thead: ({ children }) => (
          <thead style={{ background: '#2a2a2a' }}>{children}</thead>
        ),
        th: ({ children }) => (
          <th style={{
            border: '1px solid #3a3a3a',
            padding: '8px 12px',
            textAlign: 'left',
            fontWeight: 600,
            color: '#ddd',
          }}>
            {children}
          </th>
        ),
        td: ({ children }) => (
          <td style={{
            border: '1px solid #3a3a3a',
            padding: '8px 12px',
            color: '#ccc',
          }}>
            {children}
          </td>
        ),
        // Links
        a: ({ href, children }) => (
          <a
            href={href}
            target="_blank"
            rel="noopener noreferrer"
            style={{
              color: '#6bafff',
              textDecoration: 'none',
              borderBottom: '1px solid transparent',
              transition: 'border-color 0.15s',
            }}
            onMouseEnter={e => (e.currentTarget.style.borderBottomColor = '#6bafff')}
            onMouseLeave={e => (e.currentTarget.style.borderBottomColor = 'transparent')}
          >
            {children}
          </a>
        ),
        // Horizontal rule
        hr: () => (
          <hr style={{ border: 'none', borderTop: '1px solid #333', margin: '16px 0' }} />
        ),
        // Strong & em
        strong: ({ children }) => (
          <strong style={{ color: '#f0f0f0', fontWeight: 600 }}>{children}</strong>
        ),
      }}
    >
      {content}
    </ReactMarkdown>
  )
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
          fontSize: 15,
          flex: 1,
          minWidth: 0,
          overflow: 'hidden',
        }}>
          <MarkdownContent content={content} />
        </div>
      </div>
    </div>
  )
}
