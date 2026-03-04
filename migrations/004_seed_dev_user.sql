-- Migration: 004_seed_developer_user.sql
-- Description: Inserts a super-admin/developer user for emergency access.

-- 1. Ensure a condominium exists for the dev user (using a fixed UUID for idempotency)
INSERT INTO condominio (
    id_condominio, nombre, direccion, ciudad, telefono, nit, logo_url
) VALUES (
    '00000000-0000-0000-0000-000000000000',
    'Neighborhood Dev Center',
    'Calle 123 Dev',
    'Bucaramanga',
    '555-5555',
    '900000000-1',
    'https://via.placeholder.com/150'
) ON CONFLICT (id_condominio) DO NOTHING;

-- 2. Insert the Developer User
-- Pin is '999999'. 
-- NOTE: If the backend is using Bcrypt-only, this string needs to be a hash.
-- However, with the "Hybrid Auth" fix I deployed, plain text '999999' is accepted as a fallback.
INSERT INTO usuario (
    email,
    nombres,
    apellidos,
    telefono,
    celular,
    tipo_documento,
    numero_documento,
    id_condominio,
    password, -- PIN
    rol,
    estado,
    fecha_creacion
) VALUES (
    'dev@cris.ac',
    'Desarrollador',
    'SuperAdmin',
    '3001234567',
    '3001234567',
    'CC',
    '1000000000',
    '00000000-0000-0000-0000-000000000000', -- Linked to Dev Condo
    '999999', -- Plain text PIN (Supported by Hybrid Auth)
    'administrador',
    'activo',
    NOW()
) ON CONFLICT (email) DO UPDATE SET 
    password = '999999',
    rol = 'administrador',
    estado = 'activo';
