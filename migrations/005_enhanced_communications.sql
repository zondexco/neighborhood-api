-- 005_enhanced_communications.sql
-- Mejora del sistema de comunicados: programación, segmentación por rol,
-- iconos, comentarios y tracking de lectura.

-- 1. Nuevas columnas en comunicado
ALTER TABLE comunicado
  ADD COLUMN IF NOT EXISTS programado_para     TIMESTAMPTZ DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS roles_destino       TEXT[]      DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS icono               TEXT        NOT NULL DEFAULT 'megaphone',
  ADD COLUMN IF NOT EXISTS permite_comentarios BOOLEAN     NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS publicado           BOOLEAN     NOT NULL DEFAULT true;

-- 2. Tabla de lectura (tracking de "leído")
CREATE TABLE IF NOT EXISTS comunicado_lectura (
  id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  id_comunicado  UUID        NOT NULL REFERENCES comunicado(id_comunicado) ON DELETE CASCADE,
  id_usuario     UUID        NOT NULL REFERENCES usuario(id_usuario)      ON DELETE CASCADE,
  leido_en       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(id_comunicado, id_usuario)
);

CREATE INDEX IF NOT EXISTS idx_com_lectura_usuario    ON comunicado_lectura(id_usuario);
CREATE INDEX IF NOT EXISTS idx_com_lectura_comunicado ON comunicado_lectura(id_comunicado);

-- 3. Tabla de comentarios
CREATE TABLE IF NOT EXISTS comunicado_comentario (
  id_comentario  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  id_comunicado  UUID        NOT NULL REFERENCES comunicado(id_comunicado) ON DELETE CASCADE,
  id_usuario     UUID        NOT NULL REFERENCES usuario(id_usuario)      ON DELETE CASCADE,
  contenido      TEXT        NOT NULL,
  fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_com_comentario_comunicado ON comunicado_comentario(id_comunicado);
