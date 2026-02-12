-- Add valor_base to reserva and backfill from costo_total
ALTER TABLE reserva
ADD COLUMN IF NOT EXISTS valor_base DOUBLE PRECISION NULL;

-- Backfill existing rows so reporting remains consistent
UPDATE reserva
SET valor_base = costo_total
WHERE valor_base IS NULL;
