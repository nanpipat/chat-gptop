'use client'

import { useState, useEffect } from 'react'
import type { Project, FileNode } from '@/lib/api'
import { createProject, deleteProject, getFiles, getProjects } from '@/lib/api'
import FileTree from './FileTree'
import UploadButton from './UploadButton'

interface Props {
  open: boolean
  onClose: () => void
  selectedProjectIds: string[]
  onToggleProject: (id: string) => void
}

export default function ProjectPanel({ open, onClose, selectedProjectIds, onToggleProject }: Props) {
  const [projects, setProjects] = useState<Project[]>([])
  const [newName, setNewName] = useState('')
  const [expandedProject, setExpandedProject] = useState<string | null>(null)
  const [files, setFiles] = useState<FileNode[]>([])

  useEffect(() => {
    if (open) loadProjects()
  }, [open])

  async function loadProjects() {
    const data = await getProjects()
    setProjects(data)
  }

  async function handleCreate() {
    if (!newName.trim()) return
    await createProject(newName.trim())
    setNewName('')
    await loadProjects()
  }

  async function handleDelete(e: React.MouseEvent, id: string) {
    e.stopPropagation()
    if (!confirm('Delete this project and all its files?')) return
    await deleteProject(id)
    await loadProjects()
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

  if (!open) return null

  return (
    <>
      {/* Backdrop */}
      <div
        onClick={onClose}
        style={{
          position: 'fixed',
          inset: 0,
          background: 'rgba(0, 0, 0, 0.5)',
          zIndex: 40,
          animation: 'fadeIn 0.2s ease',
        }}
      />

      {/* Panel */}
      <div style={{
        position: 'fixed',
        top: 0,
        right: 0,
        bottom: 0,
        width: 420,
        maxWidth: '90vw',
        background: '#1a1a1a',
        color: '#e0e0e0',
        zIndex: 50,
        display: 'flex',
        flexDirection: 'column',
        animation: 'slideIn 0.25s ease',
        boxShadow: '-4px 0 24px rgba(0, 0, 0, 0.4)',
      }}>
        {/* Header */}
        <div style={{
          padding: '20px 20px 16px',
          borderBottom: '1px solid #2a2a2a',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}>
          <h2 style={{ fontSize: 18, fontWeight: 600, margin: 0 }}>Projects</h2>
          <button
            onClick={onClose}
            style={{
              background: 'none',
              border: 'none',
              color: '#888',
              cursor: 'pointer',
              fontSize: 20,
              padding: '4px 8px',
              borderRadius: 6,
              lineHeight: 1,
            }}
            onMouseEnter={e => (e.currentTarget.style.color = '#fff')}
            onMouseLeave={e => (e.currentTarget.style.color = '#888')}
          >
            ×
          </button>
        </div>

        {/* Create project */}
        <div style={{ padding: '16px 20px', borderBottom: '1px solid #2a2a2a' }}>
          <div style={{ display: 'flex', gap: 8 }}>
            <input
              value={newName}
              onChange={e => setNewName(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && handleCreate()}
              placeholder="New project name..."
              style={{
                flex: 1,
                background: '#252525',
                border: '1px solid #333',
                borderRadius: 8,
                color: '#e0e0e0',
                padding: '8px 12px',
                fontSize: 14,
                outline: 'none',
              }}
            />
            <button
              onClick={handleCreate}
              style={{
                background: '#d4a574',
                color: '#1a1a1a',
                border: 'none',
                borderRadius: 8,
                padding: '8px 16px',
                cursor: 'pointer',
                fontSize: 14,
                fontWeight: 600,
              }}
              onMouseEnter={e => (e.currentTarget.style.background = '#c49060')}
              onMouseLeave={e => (e.currentTarget.style.background = '#d4a574')}
            >
              Create
            </button>
          </div>
        </div>

        {/* Project list */}
        <div style={{ flex: 1, overflowY: 'auto' }}>
          {projects.length === 0 && (
            <div style={{ padding: 20, color: '#666', fontSize: 14, textAlign: 'center' }}>
              No projects yet. Create one above.
            </div>
          )}
          {projects.map(p => (
            <div key={p.id} style={{ borderBottom: '1px solid #222' }}>
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  padding: '12px 20px',
                  cursor: 'pointer',
                  background: expandedProject === p.id ? '#222' : 'transparent',
                  transition: 'background 0.15s',
                }}
                onMouseEnter={e => {
                  if (expandedProject !== p.id) e.currentTarget.style.background = '#1e1e1e'
                }}
                onMouseLeave={e => {
                  if (expandedProject !== p.id) e.currentTarget.style.background = 'transparent'
                }}
              >
                <input
                  type="checkbox"
                  checked={selectedProjectIds.includes(p.id)}
                  onChange={() => onToggleProject(p.id)}
                  style={{
                    marginRight: 12,
                    width: 16,
                    height: 16,
                    cursor: 'pointer',
                    accentColor: '#d4a574',
                  }}
                />
                <span
                  onClick={() => toggleExpand(p.id)}
                  style={{
                    flex: 1,
                    fontSize: 14,
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                    fontWeight: 500,
                  }}
                >
                  {expandedProject === p.id ? '▾' : '▸'}{' '}
                  {p.name}
                </span>
                <button
                  onClick={(e) => handleDelete(e, p.id)}
                  style={{
                    background: 'none',
                    border: 'none',
                    color: '#666',
                    cursor: 'pointer',
                    fontSize: 14,
                    padding: '2px 6px',
                    borderRadius: 4,
                  }}
                  onMouseEnter={e => (e.currentTarget.style.color = '#ef4444')}
                  onMouseLeave={e => (e.currentTarget.style.color = '#666')}
                >
                  ×
                </button>
              </div>
              {expandedProject === p.id && (
                <div style={{ padding: '8px 20px 16px 48px', background: '#1a1a1a' }}>
                  <UploadButton projectId={p.id} onUploaded={refreshFiles} />
                  <div style={{ marginTop: 10 }}>
                    {files.length === 0 ? (
                      <div style={{ color: '#555', fontSize: 13, padding: '8px 0' }}>
                        No files uploaded yet
                      </div>
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
    </>
  )
}
