-- Paquetes / Encomiendas
CREATE TABLE IF NOT EXISTS paquete (
    id_paquete UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_apartamento UUID NOT NULL REFERENCES apartamento(id_apartamento),
    residente TEXT NOT NULL,
    transportadora TEXT NOT NULL,
    notas TEXT NULL,
    fecha_recibido TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    fecha_entregado TIMESTAMP WITH TIME ZONE NULL,
    id_condominio UUID NOT NULL REFERENCES condominio(id_condominio),
    fecha_creacion TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    fecha_actualizacion TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_paquete_condominio ON paquete(id_condominio);
CREATE INDEX IF NOT EXISTS idx_paquete_entregado ON paquete(fecha_entregado);
