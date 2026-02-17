# EXACT IMPLEMENTATION PROMPT FOR WINDSURF

# Project: ChatGPT-like Multi-Project RAG System

# Stack: Go + Postgres pgvector + OpenAI API + Next.js + Docker Compose

You are a senior software engineer. Implement a complete production-ready MVP exactly according to this specification.

The system must run successfully using:

docker compose up --build

and be fully functional.

---

# OVERVIEW

Build a ChatGPT-like system with:

- Global chat (not tied to a single project)
- Multiple projects as knowledge sources
- Folder upload (entire repository)
- File tree view
- File and folder deletion
- RAG using OpenAI embeddings
- Streaming chat responses
- Next.js frontend
- Go backend
- Postgres with pgvector
- Docker Compose deployment

---

# PROJECT STRUCTURE

Create this exact folder structure:

```
rag-chat-system/

  docker-compose.yml
  .env

  backend/
    Dockerfile
    go.mod
    go.sum

    cmd/server/main.go

    internal/

      config/
        config.go

      database/
        db.go
        migration.go

      models/
        project.go
        file.go
        chunk.go
        chat.go
        message.go

      repositories/
        project_repo.go
        file_repo.go
        chunk_repo.go
        chat_repo.go
        message_repo.go

      services/
        openai_service.go
        embedding_service.go
        rag_service.go
        ingest_service.go
        file_service.go
        chat_service.go

      handlers/
        project_handler.go
        file_handler.go
        chat_handler.go

      rag/
        chunker.go
        prompt.go

  frontend/

    Dockerfile
    package.json
    next.config.js
    tsconfig.json

    app/
      layout.tsx
      page.tsx
      chat/[chatId]/page.tsx

    components/
      ChatSidebar.tsx
      ChatWindow.tsx
      MessageBubble.tsx
      ProjectSidebar.tsx
      FileTree.tsx
      UploadButton.tsx

    lib/
      api.ts

  storage/
```

---

# ENV FILE

.env

```
OPENAI_API_KEY=YOUR_KEY_HERE

DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=ragdb

STORAGE_PATH=/app/storage

BACKEND_PORT=8080
FRONTEND_PORT=3000
```

---

# DOCKER COMPOSE

docker-compose.yml

```
version: '3.9'

services:

  postgres:
    image: pgvector/pgvector:pg16
    restart: always
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: ragdb
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  backend:
    build: ./backend
    restart: always
    env_file:
      - .env
    volumes:
      - ./storage:/app/storage
    ports:
      - "8080:8080"
    depends_on:
      - postgres

  frontend:
    build: ./frontend
    restart: always
    ports:
      - "3000:3000"
    depends_on:
      - backend

volumes:
  postgres_data:
```

---

# BACKEND REQUIREMENTS

Language: Go 1.22

Framework: Echo

Database: pgx

---

# DATABASE

On startup, automatically create tables:

projects
files
document_chunks
chats
messages

files table must support folder hierarchy:

parent_id nullable

document_chunks must use:

embedding VECTOR(1536)

Enable extension:

CREATE EXTENSION IF NOT EXISTS vector;

---

# FILE STORAGE

Store files in:

/app/storage/projects/{project_id}/...

Create folders automatically.

---

# FILE INGESTION

When uploading file or folder:

- Save file to disk
- Insert into files table
- Read content
- Chunk content (500–1000 chars, overlap 100)
- Generate embedding using OpenAI text-embedding-3-small
- Store in document_chunks

---

# RAG SEARCH

Implement function:

SearchRelevantChunks(queryEmbedding, optionalProjectIDs)

SQL must use:

ORDER BY embedding <-> $1
LIMIT 10

---

# OPENAI INTEGRATION

Use official Go OpenAI client.

Embedding model:

text-embedding-3-small

Chat model:

gpt-4o-mini

Enable streaming responses.

---

# PROMPT

Use this exact prompt template:

You are an expert software engineering assistant.

Answer ONLY using provided context.

If answer is not found, say you don't know.

Context:
{context}

Question:
{question}

Answer clearly and include file names if relevant.

---

# BACKEND API

Implement all endpoints:

Projects

POST /projects
GET /projects
DELETE /projects/{id}

Files

POST /projects/{id}/upload-file
POST /projects/{id}/upload-folder
GET /projects/{id}/files
DELETE /files/{id}

Chats

POST /chats
GET /chats
GET /chats/{id}/messages

Messages

POST /chats/{id}/messages

Body:

{
"message": "user question",
"project_ids": []
}

project_ids optional.

If empty → search all projects.

Stream response using SSE.

---

# FOLDER UPLOAD

Support multipart upload of multiple files.

Reconstruct folder structure.

---

# DELETE FILE

DELETE /files/{id}

Must delete:

- file from disk
- document_chunks
- file record

If folder → recursive delete.

---

# FRONTEND REQUIREMENTS

Framework: Next.js 14 (App Router)

Language: TypeScript

UI must look similar to ChatGPT:

Sidebar:

- chats list
- projects list

Main area:

- chat messages
- input box

---

# FRONTEND FEATURES

Must implement:

Create chat
Send message
Stream response
Create project
Upload folder
View file tree
Delete file/folder

---

# FRONTEND API CLIENT

Create lib/api.ts

Functions:

createChat()
sendMessage()
createProject()
uploadFolder()
getProjects()
getChats()

---

# STREAMING

Use EventSource or fetch streaming.

Render tokens as they arrive.

---

# BACKEND DOCKERFILE

```
FROM golang:1.22

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o server cmd/server/main.go

CMD ["./server"]
```

---

# FRONTEND DOCKERFILE

```
FROM node:20

WORKDIR /app

COPY . .

RUN npm install
RUN npm run build

CMD ["npm", "start"]
```

---

# SUCCESS REQUIREMENTS

System must successfully run:

docker compose up --build

User must be able to:

create project
upload folder repo
see file tree
create chat
ask question
receive streamed answer
delete files/folders

---

# CODE QUALITY REQUIREMENTS

Use clean architecture:

handlers → services → repositories

Code must compile without errors.

Handle errors properly.

No TODO placeholders.

Fully working implementation required.

---

# FINAL STEP

After implementation, verify system works end-to-end.

Do not skip any component.

System must be production-quality MVP.

END OF SPECIFICATION.
