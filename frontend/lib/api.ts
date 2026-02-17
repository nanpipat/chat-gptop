const API_BASE = typeof window !== 'undefined'
  ? 'http://localhost:8080'
  : 'http://backend:8080';

export interface Project {
  id: string;
  name: string;
  created_at: string;
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
  await fetch(`${API_BASE}/projects/${projectId}/upload-folder`, {
    method: 'POST',
    body: form,
  });
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

export async function getChats(): Promise<Chat[]> {
  const res = await fetch(`${API_BASE}/chats`);
  return res.json();
}

export async function createChat(title?: string): Promise<Chat> {
  const res = await fetch(`${API_BASE}/chats`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title: title || '' }),
  });
  return res.json();
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
