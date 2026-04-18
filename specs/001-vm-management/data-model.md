# Data Model: Web-Based VM Management Platform

## Entities

### Platform User
Represents an authenticated dashboard user.
- `id` (UUID, Primary Key)
- `username` (String, Unique)
- `passwordHash` (String)
- `role` (Enum: `ADMIN`, `VIEWER`)
- `createdAt` (Timestamp)

### Virtual Machine (VM)
Target server payload.
- `id` (UUID, Primary Key)
- `name` (String)
- `host` (String - IP or Hostname)
- `port` (Int - default 22)
- `username` (String - SSH user)
- `authType` (Enum: `PASSWORD`, `KEY`)
- `encryptedSecret` (String - AES-256-GCM encrypted payload storing password or SSH Private Key)
- `tags` (JSON Array of Strings)
- `createdAt` (Timestamp)

### Monitored Service
A daemon mapped to a specific VM.
- `id` (UUID, Primary Key)
- `vmId` (UUID, Foreign Key)
- `serviceName` (String - e.g. `nginx.service`)
- `friendlyName` (String - e.g. "Web Server")
- `createdAt` (Timestamp)

### Audit Log Entry
A secure, immutable record of any credential access attempt.
- `id` (UUID, Primary Key)
- `userId` (UUID, Foreign Key to Platform User)
- `vmId` (UUID, Foreign Key to Virtual Machine)
- `timestamp` (Timestamp)
- `outcome` (Enum: `SUCCESS`, `FAILURE`)


## Real-Time Payloads (No persistent storage required)

### Telemetry Record
- `vmId` (UUID)
- `cpuPercent` (Float)
- `ramPercent` (Float)
- `ramUsed` (Int - MB)
- `networkTx` (Int - B/s)
- `networkRx` (Int - B/s)
- `timestamp` (Timestamp)
- `status` (Enum: `ONLINE`, `OFFLINE`, `UNREACHABLE`)
