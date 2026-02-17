'use client'

import { useState } from 'react'
import type { Project, FileNode } from '@/lib/api'
import { createProject, deleteProject, getFiles } from '@/lib/api'
import FileTree from './FileTree'
import UploadButton from './UploadButton'

interface Props {
  projects: Project[]
  selectedProjectIds: string[]
  onToggleProject: (id: string) => void
  onProjectsChanged: () => void
}

export default function ProjectSidebar({ projects, selectedProjectIds, onToggleProject, onProjectsChanged }: Props) {
  const [newName, setNewName] = useState('')
  const [expandedProject, setExpandedProject] = useState<string | null>(null)
  const [files, setFiles] = useState<FileNode[]>([])

  async function handleCreate() {
    if (!newName.trim()) return
    await createProject(newName.trim())
    setNewName('')
    onProjectsChanged()
  }

  async function handleDelete(e: React.MouseEvent, id: string) {
    e.stopPropagation()
    if (!confirm('Delete this project and all its files?')) return
    await deleteProject(id)
    onProjectsChanged()
  }

  async function toggleExpand(id: string) {
    if (expandedProject === id) {
      setExpandedProject(null)
      setFiles([])
    } else {
      setExpandedProject(id)
      const data = await getFiles(id)
      setFiles(data)
    }
  }

  async function refreshFiles() {
    if (expandedProject) {
      const data = await getFiles(expandedProject)
      setFiles(data)
    }
  }

  return (
    <div style={{
      width: 300,
      background: '#2a2b32',
      color: '#fff',
      display: 'flex',
      flexDirection: 'column',
      borderRight: '1px solid #4d4d4f',
      flexShrink: 0,
    }}>
      <div style={{ padding: 12, borderBottom: '1px solid #4d4d4f' }}>
        <div style={{ fontSize: 14, fontWeight: 600, marginBottom: 8 }}>Projects</div>
        <div style={{ display: 'flex', gap: 6 }}>
          <input
            value={newName}
            onChange={e => setNewName(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && handleCreate()}
            placeholder="New project..."
            style={{
              flex: 1,
              background: '#40414f',
              border: '1px solid #565869',
              borderRadius: 6,
              color: '#fff',
              padding: '6px 10px',
              fontSize: 13,
              outline: 'none',
            }}
          />
          <button onClick={handleCreate} style={{
            background: '#10a37f',
            color: '#fff',
            border: 'none',
            borderRadius: 6,
            padding: '6px 12px',
            cursor: 'pointer',
            fontSize: 13,
          }}>+</button>
        </div>
      </div>
      <div style={{ flex: 1, overflowY: 'auto' }}>
        {projects.map(p => (
          <div key={p.id}>
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                padding: '8px 12px',
                cursor: 'pointer',
                background: expandedProject === p.id ? '#40414f' : 'transparent',
                borderBottom: '1px solid #3a3b42',
              }}
            >
              <input
                type="checkbox"
                checked={selectedProjectIds.includes(p.id)}
                onChange={() => onToggleProject(p.id)}
                style={{ marginRight: 8 }}
              />
              <span
                onClick={() => toggleExpand(p.id)}
                style={{ flex: 1, fontSize: 13, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}
              >
                {p.name}
              </span>
              <button
                onClick={(e) => handleDelete(e, p.id)}
                style={{
                  background: 'none',
                  border: 'none',
                  color: '#ef4444',
                  cursor: 'pointer',
                  fontSize: 12,
                  opacity: 0.6,
                  padding: '2px 4px',
                }}
              >
                {'\u2715'}
              </button>
            </div>
            {expandedProject === p.id && (
              <div style={{ padding: '8px 12px', background: '#1e1f25' }}>
                <UploadButton projectId={p.id} onUploaded={refreshFiles} />
                <div style={{ marginTop: 8 }}>
                  {files.length === 0 ? (
                    <div style={{ color: '#666', fontSize: 12 }}>No files uploaded</div>
                  ) : (
                    <FileTree files={files} onRefresh={refreshFiles} />
                  )}
                </div>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
