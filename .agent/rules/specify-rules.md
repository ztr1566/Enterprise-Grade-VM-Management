# new_project Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-04-27

## Active Technologies
- Go (Golang) v1.21+, TypeScript (Strict) + React, Tailwind CSS, xterm.js (Frontend) / gorilla/websocket, golang.org/x/crypto/ssh, mattn/go-sqlite3 (Backend) (001-vm-management)
- Go 1.25 (backend), TypeScript 5.x (frontend) + `golang.org/x/crypto/ssh`, `gorilla/websocket`, `uber-go/zap`, `mattn/go-sqlite3`, React 18, Vite, Recharts, xterm.js (003-v2-enterprise-upgrade)
- SQLite (existing `vm-manager.db`, extended with `003_provisioning.sql` migration) (003-v2-enterprise-upgrade)
- Go 1.25.0 + gRPC, Protocol Buffers, `crypto/ecdsa`, `crypto/tls`, `github.com/shirou/gopsutil/v3` (004-telemetry-hardening)
- Disk-backed WAL (Agent, 50MB limit), SQLite (Backend) (004-telemetry-hardening)
- TypeScript 5.5 (strict mode), React 18.3, ES2020 target + Vite 8.0 (bundler), React Router DOM 6.30, Axios 1.15, Lucide React 1.8, Recharts 3.8, xterm 5.3 (005-vm-dashboard-ui-sync)
- N/A (frontend SPA; backend uses SQLite) (005-vm-dashboard-ui-sync)

- TypeScript (Strict), Node.js v18+ + React, Tailwind CSS, xterm.js (Frontend) / Express, Node 'ws', ssh2, SQLite3, node-crypto (Backend) (001-vm-management)

## Project Structure

```text
src/
tests/
```

## Commands

npm test && npm run lint

## Code Style

TypeScript (Strict), Node.js v18+: Follow standard conventions

## Recent Changes
- 005-vm-dashboard-ui-sync: Added TypeScript 5.5 (strict mode), React 18.3, ES2020 target + Vite 8.0 (bundler), React Router DOM 6.30, Axios 1.15, Lucide React 1.8, Recharts 3.8, xterm 5.3
- 005-vm-dashboard-ui-sync: Added TypeScript 5.5 (strict mode), React 18.3, ES2020 target + Vite 8.0 (bundler), React Router DOM 6.30, Axios 1.15, Lucide React 1.8, Recharts 3.8, xterm 5.3
- 004-telemetry-hardening: Added Go 1.25.0 + gRPC, Protocol Buffers, `crypto/ecdsa`, `crypto/tls`, `github.com/shirou/gopsutil/v3`


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
