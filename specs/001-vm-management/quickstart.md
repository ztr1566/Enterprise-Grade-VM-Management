# Quickstart Guide

This document assists developers with standing up the Web-Based VM Management application locally.

## Prerequisites
- **Go**: v1.21 or newer
- **Node.js**: v18+ (for frontend Vite build tools)
- **CGO** enabled for `mattn/go-sqlite3`

## Backend Setup (Golang)

1. Navigate to the backend directory:
   ```bash
   cd backend
   ```
2. Download dependencies:
   ```bash
   go mod download
   ```
3. Run the Go server (Listens on `:8080`, WebSockets on `/ws`):
   ```bash
   go run cmd/server/main.go
   ```

## Frontend Setup (React + Tailwind + xterm.js)

1. Open a new terminal and navigate to the frontend directory:
   ```bash
   cd frontend
   ```
2. Install dependencies:
   ```bash
   npm install
   ```
3. Start the Vite development server (Listens on `:5173`):
   ```bash
   npm run dev
   ```
4. Open the browser to `http://localhost:5173`. By default, the admin credentials are `admin` / `admin`.
