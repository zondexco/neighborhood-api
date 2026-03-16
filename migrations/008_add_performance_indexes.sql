-- Migration 008: Add missing indexes for performance at 500 rps
-- All indexes use IF NOT EXISTS and CONCURRENTLY where possible.

-- ═══════════════════════════════════════════════════════════════
-- usuario: login (FindByEmailOnly) does WHERE email = $1 → seq scan
-- ═══════════════════════════════════════════════════════════════
CREATE INDEX IF NOT EXISTS idx_usuario_email
    ON usuario (email);

-- usuario: every admin endpoint filters by condominio
CREATE INDEX IF NOT EXISTS idx_usuario_condominio
    ON usuario (id_condominio);

-- ═══════════════════════════════════════════════════════════════
-- comunicado: ListVisible filters by (id_condominio, publicado)
-- ═══════════════════════════════════════════════════════════════
CREATE INDEX IF NOT EXISTS idx_comunicado_condominio_publicado
    ON comunicado (id_condominio, publicado);

-- ═══════════════════════════════════════════════════════════════
-- reserva: all list endpoints filter by condominio
-- ═══════════════════════════════════════════════════════════════
CREATE INDEX IF NOT EXISTS idx_reserva_condominio
    ON reserva (id_condominio);

-- reserva: overlap check uses (id_espacio, fecha_inicio, fecha_fin)
CREATE INDEX IF NOT EXISTS idx_reserva_espacio_fechas
    ON reserva (id_espacio, fecha_inicio, fecha_fin);

-- reserva: ListByUsuario filters by (id_usuario, id_condominio)
CREATE INDEX IF NOT EXISTS idx_reserva_usuario
    ON reserva (id_usuario, id_condominio);

-- reserva: ListByApartment filters by (id_apartamento, id_condominio)
CREATE INDEX IF NOT EXISTS idx_reserva_apartamento
    ON reserva (id_apartamento, id_condominio);

-- ═══════════════════════════════════════════════════════════════
-- factura: every query filters by id_condominio
-- ═══════════════════════════════════════════════════════════════
CREATE INDEX IF NOT EXISTS idx_factura_condominio
    ON factura (id_condominio);

-- factura: GetByUsuario filters by (id_condominio, id_usuario)
CREATE INDEX IF NOT EXISTS idx_factura_usuario
    ON factura (id_condominio, id_usuario);

-- factura: GetByApartment filters by (id_apartamento, id_condominio)
CREATE INDEX IF NOT EXISTS idx_factura_apartamento
    ON factura (id_apartamento, id_condominio);

-- ═══════════════════════════════════════════════════════════════
-- espacio_comun: list endpoints filter by id_condominio
-- ═══════════════════════════════════════════════════════════════
CREATE INDEX IF NOT EXISTS idx_espacio_condominio
    ON espacio_comun (id_condominio);

-- ═══════════════════════════════════════════════════════════════
-- paquete: ListWithFilters often filters by id_apartamento
-- ═══════════════════════════════════════════════════════════════
CREATE INDEX IF NOT EXISTS idx_paquete_apartamento
    ON paquete (id_apartamento);

-- ═══════════════════════════════════════════════════════════════
-- apartamento: list endpoints filter by id_condominio
-- ═══════════════════════════════════════════════════════════════
CREATE INDEX IF NOT EXISTS idx_apartamento_condominio
    ON apartamento (id_condominio);
