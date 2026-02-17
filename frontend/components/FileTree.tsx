'use client'

import { useState } from 'react'
import type { FileNode } from '@/lib/api'
import { deleteFile } from '@/lib/api'

interface Props {
  files: FileNode[]
  onRefresh: () => void
  depth?: number
}

export default function FileTree({ files, onRefresh, depth = 0 }: Props) {
  return (
    <div>
      {files.map(file => (
        <FileTreeNode key={file.id} file={file} onRefresh={onRefresh} depth={depth} />
      ))}
    </div>
  )
}

function FileTreeNode({ file, onRefresh, depth }: { file: FileNode; onRefresh: () => void; depth: number }) {
  const [expanded, setExpanded] = useState(false)
  const [deleting, setDeleting] = useState(false)

  async function handleDelete(e: React.MouseEvent) {
    e.stopPropagation()
    if (!confirm(`Delete ${file.name}?`)) return
    setDeleting(true)
    await deleteFile(file.id)
    onRefresh()
  }

  return (
    <div>
      <div
        onClick={() => file.is_dir && setExpanded(!expanded)}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 6,
          padding: '5px 8px',
          paddingLeft: depth * 18 + 8,
          cursor: file.is_dir ? 'pointer' : 'default',
          fontSize: 13,
          color: '#c0c0c0',
          borderRadius: 6,
          transition: 'background 0.1s',
        }}
        onMouseEnter={e => (e.currentTarget.style.background = '#2a2a2a')}
        onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
      >
        <span style={{ width: 16, textAlign: 'center', color: '#888', fontSize: 12 }}>
          {file.is_dir ? (expanded ? '▾' : '▸') : '·'}
        </span>
        <span style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {file.name}
        </span>
        <button
          onClick={handleDelete}
          disabled={deleting}
          style={{
            background: 'none',
            border: 'none',
            color: '#555',
            cursor: 'pointer',
            fontSize: 13,
            padding: '2px 4px',
            borderRadius: 4,
          }}
          onMouseEnter={e => (e.currentTarget.style.color = '#ef4444')}
          onMouseLeave={e => (e.currentTarget.style.color = '#555')}
        >
          {deleting ? '...' : '×'}
        </button>
      </div>
      {file.is_dir && expanded && file.children && (
        <FileTree files={file.children} onRefresh={onRefresh} depth={depth + 1} />
      )}
    </div>
  )
}
