'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import type { Project, FileNode, GitConfig } from '@/lib/api'
import { createProject, deleteProject, getFiles, getProjects, getGitConfig, saveGitConfig, syncGit, removeGitConfig } from '@/lib/api'
import FileTree from './FileTree'
import UploadButton from './UploadButton'

interface Props {
  collapsed: boolean
  onToggle: () => void
  selectedProjectIds: string[]
  onToggleProject: (id: string) => void
}

export default function ProjectPanel({ collapsed, onToggle, selectedProjectIds, onToggleProject }: Props) {
  const [projects, setProjects] = useState<Project[]>([])
  const [newName, setNewName] = useState('')
  const [expandedProject, setExpandedProject] = useState<string | null>(null)
  const expandedProjectRef = useRef<string | null>(null)
  const [files, setFiles] = useState<FileNode[]>([])

  // Git state
  const [gitConfig, setGitConfig] = useState<GitConfig | null>(null)
  const [showGitForm, setShowGitForm] = useState(false)
  const [gitUrl, setGitUrl] = useState('')
  const [gitBranch, setGitBranch] = useState('main')
  const [gitToken, setGitToken] = useState('')
  const [syncing, setSyncing] = useState(false)
  const [syncError, setSyncError] = useState<string | null>(null)
  const [savingGit, setSavingGit] = useState(false)

  useEffect(() => {
    loadProjects()
  }, [])

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
      expandedProjectRef.current = null
      setFiles([])
      setGitConfig(null)
      setShowGitForm(false)
    } else {
      setExpandedProject(id)
      expandedProjectRef.current = id
      const data = await getFiles(id)
      setFiles(data)
      // Load git config
      const gc = await getGitConfig(id)
      setGitConfig(gc)
      setShowGitForm(false)
      setSyncError(null)
    }
  }

  const refreshFiles = useCallback(async () => {
    const pid = expandedProjectRef.current
    if (pid) {
      const data = await getFiles(pid)
      setFiles(data)
    }
  }, [])

  function openGitForm(existing?: GitConfig | null) {
    if (existing) {
      setGitUrl(existing.git_url)
      setGitBranch(existing.git_branch)
      setGitToken('')
    } else {
      setGitUrl('')
      setGitBranch('main')
      setGitToken('')
    }
    setShowGitForm(true)
    setSyncError(null)
  }

  async function handleSaveGit() {
    if (!expandedProject || !gitUrl.trim()) return
    setSavingGit(true)
    setSyncError(null)
    try {
      await saveGitConfig(expandedProject, gitUrl.trim(), gitBranch.trim() || 'main', gitToken || undefined)
      const gc = await getGitConfig(expandedProject)
      setGitConfig(gc)
      setShowGitForm(false)
      setGitToken('')
    } catch (err: any) {
      setSyncError(err.message || 'Failed to save')
    } finally {
      setSavingGit(false)
    }
  }

  async function handleSync() {
    if (!expandedProject) return
    setSyncing(true)
    setSyncError(null)
    try {
      await syncGit(expandedProject)
      // Poll for completion
      const pid = expandedProject
      const poll = setInterval(async () => {
        try {
          const gc = await getGitConfig(pid)
          if (!gc) { clearInterval(poll); setSyncing(false); return }
          setGitConfig(gc)
          if (gc.sync_status === 'done') {
            clearInterval(poll)
            setSyncing(false)
            await refreshFiles()
          } else if (gc.sync_status === 'error') {
            clearInterval(poll)
            setSyncing(false)
            setSyncError(gc.sync_error || 'Sync failed')
          }
          // still 'syncing' — keep polling
        } catch {
          clearInterval(poll)
          setSyncing(false)
        }
      }, 3000)
    } catch (err: any) {
      setSyncError(err.message || 'Sync failed')
      setSyncing(false)
    }
  }

  async function handleRemoveGit() {
    if (!expandedProject) return
    if (!confirm('Unlink Git repository? Files will remain.')) return
    await removeGitConfig(expandedProject)
    setGitConfig(null)
    setShowGitForm(false)
  }

  const inputStyle: React.CSSProperties = {
    width: '100%',
    background: '#252525',
    border: '1px solid #333',
    borderRadius: 6,
    color: '#e0e0e0',
    padding: '6px 10px',
    fontSize: 12,
    outline: 'none',
    marginBottom: 6,
  }

  const smallBtnStyle: React.CSSProperties = {
    background: 'none',
    border: '1px solid #444',
    borderRadius: 6,
    color: '#ccc',
    cursor: 'pointer',
    fontSize: 11,
    padding: '4px 10px',
    transition: 'all 0.15s',
  }

  return (
    <div
      className="sidebar-transition"
      style={{
        width: collapsed ? 48 : 360,
        minWidth: collapsed ? 48 : 360,
        background: '#1a1a1a',
        color: '#e0e0e0',
        display: 'flex',
        flexDirection: 'column',
        height: '100vh',
        borderLeft: '1px solid #2a2a2a',
      }}
    >
      {/* Toggle button */}
      <div style={{
        padding: collapsed ? '12px 0' : '12px 16px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: collapsed ? 'center' : 'space-between',
        borderBottom: collapsed ? 'none' : '1px solid #2a2a2a',
      }}>
        <button
          onClick={onToggle}
          title={collapsed ? 'Open projects' : 'Close projects'}
          style={{
            background: 'none',
            border: 'none',
            color: selectedProjectIds.length > 0 ? '#d4a574' : '#888',
            cursor: 'pointer',
            fontSize: 18,
            padding: '6px 8px',
            borderRadius: 6,
            lineHeight: 1,
            flexShrink: 0,
            transition: 'color 0.15s, background 0.15s',
          }}
          onMouseEnter={e => {
            e.currentTarget.style.color = '#d4a574'
            e.currentTarget.style.background = '#2a2a2a'
          }}
          onMouseLeave={e => {
            e.currentTarget.style.color = selectedProjectIds.length > 0 ? '#d4a574' : '#888'
            e.currentTarget.style.background = 'none'
          }}
        >
          📂
        </button>

        {!collapsed && (
          <span style={{ fontSize: 15, fontWeight: 600 }}>Projects</span>
        )}

        {!collapsed && (
          <span style={{ width: 32 }} /> 
        )}
      </div>

      {/* Content - hidden when collapsed */}
      <div
        className={`sidebar-content ${collapsed ? 'collapsed' : ''}`}
        style={{
          flex: 1,
          overflowY: 'auto',
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        {!collapsed && (
          <>
            {/* Create project */}
            <div style={{ padding: '12px 16px', borderBottom: '1px solid #2a2a2a' }}>
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
                    fontSize: 13,
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
                    padding: '8px 14px',
                    cursor: 'pointer',
                    fontSize: 13,
                    fontWeight: 600,
                    transition: 'background 0.15s',
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
                      padding: '10px 16px',
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
                        marginRight: 10,
                        width: 15,
                        height: 15,
                        cursor: 'pointer',
                        accentColor: '#d4a574',
                      }}
                    />
                    <span
                      onClick={() => toggleExpand(p.id)}
                      style={{
                        flex: 1,
                        fontSize: 13,
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
                        transition: 'color 0.15s',
                      }}
                      onMouseEnter={e => (e.currentTarget.style.color = '#ef4444')}
                      onMouseLeave={e => (e.currentTarget.style.color = '#666')}
                    >
                      ×
                    </button>
                  </div>
                  {expandedProject === p.id && (
                    <div style={{ padding: '8px 16px 14px 42px', background: '#1a1a1a' }}>
                      <UploadButton projectId={p.id} onUploaded={refreshFiles} />

                      {/* Git Sync Section */}
                      <div style={{
                        marginTop: 10,
                        padding: '8px 10px',
                        background: '#222',
                        borderRadius: 8,
                        border: '1px solid #2a2a2a',
                      }}>
                        {!gitConfig && !showGitForm && (
                          <button
                            onClick={() => openGitForm()}
                            style={{
                              ...smallBtnStyle,
                              width: '100%',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              gap: 6,
                              padding: '6px 10px',
                            }}
                            onMouseEnter={e => {
                              e.currentTarget.style.borderColor = '#d4a574'
                              e.currentTarget.style.color = '#d4a574'
                            }}
                            onMouseLeave={e => {
                              e.currentTarget.style.borderColor = '#444'
                              e.currentTarget.style.color = '#ccc'
                            }}
                          >
                            🔗 Link Git Repo
                          </button>
                        )}

                        {showGitForm && (
                          <div>
                            <div style={{ fontSize: 11, color: '#999', marginBottom: 6, fontWeight: 600 }}>
                              Git Repository
                            </div>
                            <input
                              value={gitUrl}
                              onChange={e => setGitUrl(e.target.value)}
                              placeholder="https://github.com/user/repo.git"
                              style={inputStyle}
                            />
                            <input
                              value={gitBranch}
                              onChange={e => setGitBranch(e.target.value)}
                              placeholder="Branch (default: main)"
                              style={inputStyle}
                            />
                            <input
                              type="password"
                              value={gitToken}
                              onChange={e => setGitToken(e.target.value)}
                              placeholder="PAT token (optional, for private repos)"
                              style={inputStyle}
                            />
                            <div style={{ display: 'flex', gap: 6, marginTop: 2 }}>
                              <button
                                onClick={handleSaveGit}
                                disabled={savingGit || !gitUrl.trim()}
                                style={{
                                  ...smallBtnStyle,
                                  background: '#d4a574',
                                  color: '#1a1a1a',
                                  border: 'none',
                                  fontWeight: 600,
                                  opacity: savingGit || !gitUrl.trim() ? 0.5 : 1,
                                }}
                              >
                                {savingGit ? 'Saving...' : 'Save'}
                              </button>
                              <button
                                onClick={() => setShowGitForm(false)}
                                style={smallBtnStyle}
                              >
                                Cancel
                              </button>
                            </div>
                          </div>
                        )}

                        {gitConfig && !showGitForm && (
                          <div>
                            <div style={{ fontSize: 11, color: '#999', marginBottom: 4, fontWeight: 600 }}>
                              🔗 Git Linked
                            </div>
                            <div style={{ fontSize: 11, color: '#bbb', marginBottom: 2, wordBreak: 'break-all' }}>
                              {gitConfig.git_url}
                            </div>
                            <div style={{ fontSize: 10, color: '#777', marginBottom: 2 }}>
                              Branch: {gitConfig.git_branch}
                              {gitConfig.has_token && ' • 🔑 Token set'}
                            </div>
                            {gitConfig.last_synced_at && (
                              <div style={{ fontSize: 10, color: '#666', marginBottom: 6 }}>
                                Last synced: {new Date(gitConfig.last_synced_at).toLocaleString()}
                              </div>
                            )}

                            {syncError && (
                              <div style={{
                                fontSize: 11,
                                color: '#ef4444',
                                padding: '4px 8px',
                                background: '#2a1515',
                                borderRadius: 4,
                                marginBottom: 6,
                                wordBreak: 'break-word',
                              }}>
                                {syncError}
                              </div>
                            )}

                            <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
                              <button
                                onClick={handleSync}
                                disabled={syncing}
                                style={{
                                  ...smallBtnStyle,
                                  background: syncing ? '#333' : '#2a5a2a',
                                  color: syncing ? '#888' : '#8fdf8f',
                                  border: 'none',
                                  fontWeight: 600,
                                }}
                              >
                                {syncing ? '⏳ Syncing...' : '🔄 Sync Now'}
                              </button>
                              <button
                                onClick={() => openGitForm(gitConfig)}
                                style={smallBtnStyle}
                              >
                                ✏️ Edit
                              </button>
                              <button
                                onClick={handleRemoveGit}
                                style={{
                                  ...smallBtnStyle,
                                  color: '#999',
                                }}
                                onMouseEnter={e => (e.currentTarget.style.color = '#ef4444')}
                                onMouseLeave={e => (e.currentTarget.style.color = '#999')}
                              >
                                🗑️ Unlink
                              </button>
                            </div>
                          </div>
                        )}
                      </div>

                      <div style={{ marginTop: 8 }}>
                        {files.length === 0 ? (
                          <div style={{ color: '#555', fontSize: 12, padding: '6px 0' }}>
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
          </>
        )}
      </div>
    </div>
  )
}
