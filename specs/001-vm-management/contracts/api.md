# HTTP REST API Contract

The REST API is used strictly for CRUD operations on inventory, authentication, and stateful tasks that do not require low-latency streaming.

## Authentication
`POST /api/auth/login`
- **Request**: `{ "username": "...", "password": "..." }`
- **Response**: `200 OK` + Session Cookie (HTTPOnly)

## Virtual Machines (VM)

`GET /api/vms`
- **Response**: `200 OK`
```json
[
  {
    "id": "uuid",
    "name": "Production Web App",
    "host": "192.168.1.100",
    "tags": ["prod", "web"],
    "status": "ONLINE"
  }
]
```

`POST /api/vms`
- **Request**: 
```json
{
  "name": "Production Database",
  "host": "192.168.1.200",
  "port": 22,
  "username": "root",
  "authType": "PASSWORD",
  "secret": "my-secure-password",
  "tags": ["prod", "db"]
}
```
- **Response**: `201 Created`

## Monitored Services

`GET /api/vms/:vmId/services`
- **Response**: `200 OK`
```json
[
  {
    "id": "uuid",
    "serviceName": "nginx.service",
    "friendlyName": "Web Server"
  }
]
```

`POST /api/vms/:vmId/actions`
- **Request**: `{ "actionType": "RESTART_SERVICE", "serviceName": "nginx.service" }`
- **Response**: `202 Accepted`
