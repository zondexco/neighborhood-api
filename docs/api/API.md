# API Documentation

## Base URL

- Development: `http://localhost:8080`
- Production: `https://api.neighborhood.com`

## Authentication

All protected endpoints require JWT token in Authorization header:

```
Authorization: Bearer <access_token>
```

## Endpoints

### Auth - No Authentication Required

#### 1. Login

```
POST /api/v1/auth/login
```

**Request:**
```json
{
  "email": "user@example.com",
  "pin": "123456"
}
```

**Query Parameters:**
- `condominio_id` (required): The condominio ID

**Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "condominio_id": "550e8400-e29b-41d4-a716-446655440000",
  "expires_at": "2025-10-29T10:30:00Z"
}
```

**Error (401):**
```json
{
  "error": true,
  "message": "invalid email or PIN",
  "code": "UNAUTHORIZED"
}
```

#### 2. Refresh Token

```
POST /api/v1/auth/refresh
```

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200):** Same as Login response

#### 3. Logout

```
POST /api/v1/auth/logout
```

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200):**
```json
{
  "success": true,
  "message": "logged out successfully"
}
```

---

### Apartments - Protected Endpoints

#### 4. Get Apartments

```
GET /api/v1/apartments?page=1&page_size=10
```

**Headers:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `page` (optional, default=1): Page number
- `page_size` (optional, default=10): Items per page (max=100)

**Response (200):**
```json
{
  "total": 2,
  "page": 1,
  "page_size": 10,
  "data": [
    {
      "id": "apt-001",
      "numero": "101",
      "area": 75.5,
      "estado": "ocupado",
      "piso": 1,
      "propietario_id": "user-001",
      "propietario_nombre": "Juan Pérez",
      "condominio_id": "cond-001",
      "created_at": "2025-01-10T08:00:00Z",
      "updated_at": "2025-01-10T08:00:00Z"
    }
  ]
}
```

#### 5. Get Apartment Details

```
GET /api/v1/apartments/:id
```

**Response (200):** Single apartment object (same structure as above)

---

### Invoices - Protected Endpoints

#### 6. Get Invoices

```
GET /api/v1/facturas?page=1&page_size=10
```

**Response (200):**
```json
{
  "total": 5,
  "page": 1,
  "page_size": 10,
  "data": [
    {
      "id": "inv-001",
      "numero": "FAC-2025-0001",
      "monto": 250.00,
      "estado": "pendiente",
      "fecha_vencimiento": "2025-11-30T00:00:00Z",
      "fecha_emision": "2025-10-28T00:00:00Z",
      "descripcion_concepto": "Cuota mensual octubre",
      "apartamento_id": "apt-001",
      "apartamento_numero": "101",
      "created_at": "2025-10-28T10:00:00Z",
      "updated_at": "2025-10-28T10:00:00Z"
    }
  ]
}
```

#### 7. Get Invoice Details

```
GET /api/v1/facturas/:id
```

**Response (200):** Single invoice object

---

### Reservations - Protected Endpoints

#### 8. Get Reservations

```
GET /api/v1/reservas?page=1&page_size=10
```

**Response (200):**
```json
{
  "total": 3,
  "page": 1,
  "page_size": 10,
  "data": [
    {
      "id": "res-001",
      "espacio_id": "esp-001",
      "espacio_nombre": "Salón Comunal",
      "usuario_id": "user-001",
      "fecha_inicio": "2025-11-05T10:00:00Z",
      "fecha_fin": "2025-11-05T14:00:00Z",
      "personas_esperadas": 30,
      "estado": "confirmada",
      "condominio_id": "cond-001",
      "created_at": "2025-10-28T10:30:00Z",
      "updated_at": "2025-10-28T10:30:00Z"
    }
  ]
}
```

#### 9. Create Reservation

```
POST /api/v1/reservas
```

**Request:**
```json
{
  "espacio_id": "esp-001",
  "fecha_inicio": "2025-11-05T10:00:00Z",
  "fecha_fin": "2025-11-05T14:00:00Z",
  "personas_esperadas": 25
}
```

**Response (201):** Reservation object

#### 10. Get Reservation Details

```
GET /api/v1/reservas/:id
```

---

### Communications - Protected Endpoints

#### 11. Get Communications

```
GET /api/v1/comunicados?page=1&page_size=10
```

**Response (200):**
```json
{
  "total": 2,
  "page": 1,
  "page_size": 10,
  "data": [
    {
      "id": "com-001",
      "titulo": "Mantenimiento de Ascensores",
      "contenido": "Se realizará mantenimiento de ascensores el próximo sábado...",
      "fecha": "2025-10-28T09:00:00Z",
      "autor": "Administración",
      "condominio_id": "cond-001",
      "created_at": "2025-10-28T08:00:00Z",
      "updated_at": "2025-10-28T08:00:00Z"
    }
  ]
}
```

---

## Error Codes

| Code | Status | Description |
|------|--------|-------------|
| VALIDATION_ERROR | 400 | Invalid request format |
| BAD_REQUEST | 400 | Bad request |
| NOT_FOUND_ERROR | 404 | Resource not found |
| UNAUTHORIZED | 401 | Missing or invalid authentication |
| FORBIDDEN | 403 | Insufficient permissions |
| CONFLICT | 409 | Resource conflict |
| DATABASE_ERROR | 500 | Database error |
| INTERNAL_ERROR | 500 | Internal server error |

---

## Rate Limiting

(To be implemented in Phase 2)

---

## Pagination

Most list endpoints support pagination:

- `page`: Current page (default: 1)
- `page_size`: Items per page (default: 10, max: 100)

Response includes:
- `total`: Total number of items
- `page`: Current page
- `page_size`: Items per page
- `data`: Array of items

---

**Last Updated**: October 28, 2025
