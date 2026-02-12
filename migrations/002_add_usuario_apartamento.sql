-- Link usuarios to apartamentos
ALTER TABLE usuario
ADD COLUMN IF NOT EXISTS id_apartamento UUID NULL REFERENCES apartamento(id_apartamento);

CREATE INDEX IF NOT EXISTS idx_usuario_apartamento ON usuario(id_apartamento);
