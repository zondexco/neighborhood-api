-- 007_notificaciones.sql
-- Tabla de notificaciones in-app personales por usuario
-- Se genera automáticamente cuando: se registra un paquete o cambia el estado de una reserva

CREATE TABLE IF NOT EXISTS notificacion (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  id_usuario      UUID        NOT NULL REFERENCES usuario(id_usuario)      ON DELETE CASCADE,
  id_condominio   UUID        NOT NULL,
  tipo            VARCHAR(50) NOT NULL, -- 'paquete' | 'reserva'
  titulo          TEXT        NOT NULL,
  mensaje         TEXT        NOT NULL,
  leido           BOOLEAN     NOT NULL DEFAULT false,
  referencia_id   UUID        DEFAULT NULL, -- id del paquete o reserva
  fecha_creacion  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notificacion_usuario    ON notificacion(id_usuario);
CREATE INDEX IF NOT EXISTS idx_notificacion_condominio ON notificacion(id_condominio);
CREATE INDEX IF NOT EXISTS idx_notificacion_leido      ON notificacion(id_usuario, leido);
