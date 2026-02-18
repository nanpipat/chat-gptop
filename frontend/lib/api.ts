const API_BASE = typeof window !== 'undefined'
  ? 'http://localhost:8080'
  : 'http://backend:8080';

export interface Project {
  id: string;
  name: string;
  git_url?: string;
  git_branch?: string;
  last_synced_at?: string;
  created_at: string;
}

export interface GitConfig {
  git_url: string;
  git_branch: string;
  has_token: boolean;
  last_synced_at?: string;
  sync_status?: string;  // "syncing", "done", "error"
  sync_error?: string;
}

export interface FileNode {
  id: string;
  project_id: string;
  parent_id: string | null;
  name: string;
  path: string;
  is_dir: boolean;
  created_at: string;
  children?: FileNode[];
}

export interface Chat {
  id: string;
  title: string;
  project_ids: string[];
  created_at: string;
}

export interface Message {
  id: string;
  chat_id: string;
  role: string;
  content: string;
  created_at: string;
}

export async function getProjects(): Promise<Project[]> {
  const res = await fetch(`${API_BASE}/projects`);
  return res.json();
}

export async function createProject(name: string): Promise<Project> {
  const res = await fetch(`${API_BASE}/projects`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name }),
  });
  return res.json();
}

export async function deleteProject(id: string): Promise<void> {
  await fetch(`${API_BASE}/projects/${id}`, { method: 'DELETE' });
}

export async function getFiles(projectId: string): Promise<FileNode[]> {
  const res = await fetch(`${API_BASE}/projects/${projectId}/files`);
  return res.json();
}

export async function uploadFolder(projectId: string, files: FileList): Promise<void> {
  const form = new FormData();
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    form.append('files', file);
    const relativePath = (file as any).webkitRelativePath || file.name;
    form.append('paths', relativePath);
  }
  const res = await fetch(`${API_BASE}/projects/${projectId}/upload-folder`, {
    method: 'POST',
    body: form,
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({ error: 'Upload failed' }));
    throw new Error(data.error || `Upload failed (${res.status})`);
  }
}

export async function uploadFile(projectId: string, file: File): Promise<void> {
  const form = new FormData();
  form.append('file', file);
  await fetch(`${API_BASE}/projects/${projectId}/upload-file`, {
    method: 'POST',
    body: form,
  });
}

export async function deleteFile(fileId: string): Promise<void> {
  await fetch(`${API_BASE}/files/${fileId}`, { method: 'DELETE' });
}

// Git sync API

export async function getGitConfig(projectId: string): Promise<GitConfig | null> {
  const res = await fetch(`${API_BASE}/projects/${projectId}/git`);
  const data = await res.json();
  if (!data.git_url) return null;
  return data;
}

export async function saveGitConfig(
  projectId: string,
  gitUrl: string,
  gitBranch: string,
  token?: string,
): Promise<void> {
  const body: Record<string, string> = { git_url: gitUrl, git_branch: gitBranch };
  if (token) body.token = token;
  const res = await fetch(`${API_BASE}/projects/${projectId}/git`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({ error: 'Save failed' }));
    throw new Error(data.error || 'Save failed');
  }
}

export async function syncGit(projectId: string): Promise<void> {
  const res = await fetch(`${API_BASE}/projects/${projectId}/git/sync`, {
    method: 'POST',
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({ error: 'Sync failed' }));
    throw new Error(data.error || 'Sync failed');
  }
}

export async function removeGitConfig(projectId: string): Promise<void> {
  await fetch(`${API_BASE}/projects/${projectId}/git`, { method: 'DELETE' });
}

// Chat API

export async function getChats(): Promise<Chat[]> {
  const res = await fetch(`${API_BASE}/chats`);
  return res.json();
}

export async function deleteChat(id: string): Promise<void> {
  await fetch(`${API_BASE}/chats/${id}`, { method: 'DELETE' });
}

export async function createChat(title?: string, projectIds?: string[]): Promise<Chat> {
  const res = await fetch(`${API_BASE}/chats`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title: title || '', project_ids: projectIds || [] }),
  });
  return res.json();
}

export async function updateChatProjects(chatId: string, projectIds: string[]): Promise<void> {
  await fetch(`${API_BASE}/chats/${chatId}/projects`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ project_ids: projectIds }),
  });
}

export async function getMessages(chatId: string): Promise<Message[]> {
  const res = await fetch(`${API_BASE}/chats/${chatId}/messages`);
  return res.json();
}

export function sendMessage(
  chatId: string,
  message: string,
  projectIds: string[],
  onToken: (token: string) => void,
  onDone: () => void,
  onError: (err: string) => void,
): AbortController {
  const controller = new AbortController();

  fetch(`${API_BASE}/chats/${chatId}/messages`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ message, project_ids: projectIds }),
    signal: controller.signal,
  })
    .then(async (res) => {
      if (!res.body) {
        onError('No response body');
        return;
      }
      const reader = res.body.getReader();
      const decoder = new TextDecoder();

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        const text = decoder.decode(value, { stream: true });
        const lines = text.split('\n');

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6);
            if (data === '[DONE]') {
              onDone();
              return;
            }
            if (data.startsWith('[ERROR]')) {
              onError(data);
              return;
            }
            onToken(data);
          }
        }
      }
      onDone();
    })
    .catch((err) => {
      if (err.name !== 'AbortError') {
        onError(err.message);
      }
    });

  return controller;
}
