# Contracts: gRPC Interceptors & Identity Flow

## 1. Enrollment & Authorization (CSR)

**RPC**: `EnrollAgent(CSRRequest) returns (CertificateResponse)`

- **Metadata/Headers Requirements**:
  - `authorization`: Must contain the `Bearer <OTT>`.
- **Payload Restrictions**:
  - CSR Subject must contain `O=<organizational-unit>`.
  - SANs must contain the VM's exact IP and Hostname.
- **Backend Validation**:
  - Checks if OTT exists and is not expired (<= 10 mins).
  - Invalidates OTT immediately upon successful issuance.
  - Returns issued certificate with a strict 30-day TTL.

## 2. Telemetry Streaming (mTLS)

**RPCs**: `StreamMetrics(...)`, `StreamLogs(...)`

- **Connection Requirements**:
  - Must establish mTLS using the certificate issued during enrollment.
- **Backend Validation (Interceptor)**:
  - Validates client certificate serial number against the active SQLite CRL.
  - Rejects connection with `codes.Unauthenticated` if revoked.
- **Payload Restrictions (Interceptor)**:
  - Enforces `MaxRecvMsgSize` = 4MB.
  - Rejects payload with `codes.ResourceExhausted` if > 4MB.

## 3. Agent Health Endpoint (Local HTTP)

**Endpoint**: `GET http://127.0.0.1:9090/health` (Internal only)

**Response (JSON)**:
```json
{
  "status": "healthy",
  "memory_mb": 45.2,
  "cpu_percent": 2.1,
  "wal_state": "clean",
  "counters": {
    "ott_rejections": 0,
    "crl_rejections": 0,
    "wal_recoveries": 1,
    "wal_evictions": 0,
    "backpressure_activations": 2
  }
}
```
