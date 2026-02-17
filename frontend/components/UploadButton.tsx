'use client'

import { useRef, useState } from 'react'
import { uploadFolder, uploadFile as uploadSingleFile } from '@/lib/api'

interface Props {
  projectId: string
  onUploaded: () => void
}

export default function UploadButton({ projectId, onUploaded }: Props) {
  const folderRef = useRef<HTMLInputElement>(null)
  const fileRef = useRef<HTMLInputElement>(null)
  const [uploading, setUploading] = useState(false)

  async function handleFolderUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const files = e.target.files
    if (!files || files.length === 0) return
    setUploading(true)
    try {
      await uploadFolder(projectId, files)
      onUploaded()
    } catch (err) {
      console.error('Upload error:', err)
    }
    setUploading(false)
    if (folderRef.current) folderRef.current.value = ''
  }

  async function handleFileUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const files = e.target.files
    if (!files || files.length === 0) return
    setUploading(true)
    try {
      for (let i = 0; i < files.length; i++) {
        await uploadSingleFile(projectId, files[i])
      }
      onUploaded()
    } catch (err) {
      console.error('Upload error:', err)
    }
    setUploading(false)
    if (fileRef.current) fileRef.current.value = ''
  }

  return (
    <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
      <input
        ref={folderRef}
        type="file"
        // @ts-expect-error webkitdirectory is non-standard
        webkitdirectory=""
        multiple
        onChange={handleFolderUpload}
        style={{ display: 'none' }}
      />
      <input
        ref={fileRef}
        type="file"
        multiple
        onChange={handleFileUpload}
        style={{ display: 'none' }}
      />
      <button
        onClick={() => folderRef.current?.click()}
        disabled={uploading}
        style={uploadBtn}
        onMouseEnter={e => (e.currentTarget.style.borderColor = '#555')}
        onMouseLeave={e => (e.currentTarget.style.borderColor = '#333')}
      >
        {uploading ? 'Uploading...' : 'Upload Folder'}
      </button>
      <button
        onClick={() => fileRef.current?.click()}
        disabled={uploading}
        style={uploadBtn}
        onMouseEnter={e => (e.currentTarget.style.borderColor = '#555')}
        onMouseLeave={e => (e.currentTarget.style.borderColor = '#333')}
      >
        Upload File
      </button>
    </div>
  )
}

const uploadBtn: React.CSSProperties = {
  padding: '6px 14px',
  background: '#252525',
  color: '#c0c0c0',
  border: '1px solid #333',
  borderRadius: 6,
  cursor: 'pointer',
  fontSize: 13,
  transition: 'border-color 0.15s',
}
