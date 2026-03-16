-- Migration 009: Prevent double-booking with an EXCLUDE constraint.
-- Requires btree_gist extension to combine equality (=) with range overlap (&&).

CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Add exclusion constraint: no two non-cancelled reservations for the same space
-- can have overlapping time ranges.
-- If the constraint already exists this will error; wrapped in DO block for safety.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'excl_reserva_overlap'
    ) THEN
        ALTER TABLE reserva
            ADD CONSTRAINT excl_reserva_overlap
            EXCLUDE USING gist (
                id_espacio WITH =,
                tstzrange(fecha_inicio, fecha_fin) WITH &&
            )
            WHERE (estado NOT IN ('cancelado', 'rechazada'));
    END IF;
END $$;
