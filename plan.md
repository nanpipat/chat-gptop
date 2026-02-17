# Project Plan: Project-Scoped ChatGPT with RAG (OpenAI API, Go, Postgres, pgvector)

## Goal

Build a web application similar to ChatGPT that allows users to:

- Create projects
- Upload code repositories and documents
- Ask questions about uploaded sources
- Get answers based on project-specific knowledge using RAG
- Run locally using docker compose

This is an MVP but production-structured.

---

# Tech Stack

## Backend

- Go 1.22+
- Echo framework
- PostgreSQL 16
- pgvector extension
- OpenAI API
- sqlx or pgx
- Docker

## Frontend (MVP)

- Simple static HTML + JS (later upgrade to Nuxt/Next)

## Infrastructure

- Docker Compose
- Postgres with pgvector
- Local file storage (later replace with S3)

---

# System Architecture

```
                ┌─────────────┐
                │   Frontend  │
                │ Chat UI     │
                └──────┬──────┘
                       │ HTTP
                       ▼
                ┌─────────────┐
                │   Backend   │
                │   Go API    │
                └──────┬──────┘
                       │
     ┌─────────────────┼─────────────────┐
     ▼                 ▼                 ▼
File Storage     PostgreSQL         OpenAI API
(local fs)       pgvector          embeddings/chat
```

---

# Core Features

## Projects

- Create project
- List projects

## Files

- Upload files
- Store in filesystem
- Ingest and embed

## Chat

- Ask question
- Retrieve relevant chunks
- Call OpenAI
- Return answer

---

# Database Schema

## projects

```sql
CREATE TABLE projects (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## files

```sql
CREATE TABLE files (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id),
    filename TEXT,
    path TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## document_chunks

```sql
CREATE TABLE document_chunks (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id),
    file_id UUID REFERENCES files(id),
    content TEXT,
    embedding VECTOR(1536)
);
```

## chats

```sql
CREATE TABLE chats (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id),
    created_at TIMESTAMP DEFAULT NOW()
);
```

## messages

```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    chat_id UUID REFERENCES chats(id),
    role TEXT,
    content TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

# Project Structure

```
project-rag-chat/
│
├ backend/
│  ├ main.go
│  ├ config/
│  │   config.go
│  │
│  ├ database/
│  │   db.go
│  │
│  ├ models/
│  │   project.go
│  │   file.go
│  │   chunk.go
│  │   chat.go
│  │
│  ├ handlers/
│  │   project_handler.go
│  │   upload_handler.go
│  │   chat_handler.go
│  │
│  ├ services/
│  │   embedding_service.go
│  │   rag_service.go
│  │   openai_service.go
│  │   file_service.go
│  │
│  └ utils/
│      chunker.go
│
├ frontend/
│  └ index.html
│
├ storage/
│
├ docker-compose.yml
├ Dockerfile
├ .env
└ README.md
```

---

# Environment Variables

.env

```
OPENAI_API_KEY=your_key_here

DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=ragdb

STORAGE_PATH=/app/storage
```

---

# Docker Compose

docker-compose.yml

```
version: '3.9'

services:

  postgres:
    image: pgvector/pgvector:pg16
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: ragdb
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  backend:
    build: .
    ports:
      - "8080:8080"
    env_file:
      - .env
    volumes:
      - ./storage:/app/storage
    depends_on:
      - postgres

volumes:
  postgres_data:
```

---

# Backend Implementation Tasks

---

# Task 1: Setup Go project

Initialize:

```
go mod init project-rag-chat
```

Install:

```
go get github.com/labstack/echo/v4
go get github.com/jackc/pgx/v5
go get github.com/google/uuid
go get github.com/sashabaranov/go-openai
```

---

# Task 2: Database connection

database/db.go

Responsibilities:

- Connect postgres
- Enable pgvector
- Run migrations

---

# Task 3: Project API

Endpoints:

```
POST /projects
GET /projects
```

---

# Task 4: File Upload API

Endpoint:

```
POST /projects/:id/upload
```

Flow:

- save file
- insert file record
- read content
- chunk content
- create embeddings
- store chunks

---

# Task 5: Chunking logic

utils/chunker.go

Split text:

- 500–1000 chars per chunk
- overlap 100 chars

---

# Task 6: Embedding Service

services/embedding_service.go

Function:

```
CreateEmbedding(text string) []float32
```

Use model:

```
text-embedding-3-small
```

---

# Task 7: RAG Search Service

services/rag_service.go

Function:

```
Search(projectID, questionEmbedding)
```

SQL:

```
SELECT content
FROM document_chunks
WHERE project_id=$1
ORDER BY embedding <-> $2
LIMIT 5
```

---

# Task 8: OpenAI Chat Service

services/openai_service.go

Function:

```
Ask(context, question)
```

Prompt:

```
You are an assistant helping with a software project.

Answer ONLY using provided context.

Context:
{context}

Question:
{question}
```

Model:

```
gpt-4o-mini
```

---

# Task 9: Chat API

Endpoint:

```
POST /projects/:id/chat
```

Flow:

1 embed question
2 search chunks
3 build context
4 call OpenAI
5 return answer

---

# Task 10: Frontend MVP

frontend/index.html

Features:

- project list
- create project
- upload file
- chat box

Simple fetch API

---

# Task 11: Dockerfile

```
FROM golang:1.22

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o app backend/main.go

CMD ["./app"]
```

---

# Task 12: Run migrations automatically

On backend start:

Create tables if not exists

---

# Task 13: Run system

Start:

```
docker compose up --build
```

Backend:

```
http://localhost:8080
```

---

# API Summary

Create project:

```
POST /projects
```

Upload file:

```
POST /projects/:id/upload
```

Chat:

```
POST /projects/:id/chat

{
  "message": "Does this project support recurring tasks?"
}
```

---

# Future Improvements

- Git repo upload
- Streaming response
- Auth
- S3 storage
- Background worker
- Delete files
- Re-index

---

# MVP Success Criteria

System must:

- run via docker compose up
- create project
- upload file
- ask question
- answer using uploaded content

---

# End of Plan
