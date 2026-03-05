-- Add permite_soporte flag to condominio table
-- Allows admins to enable/disable tech support access for devs
ALTER TABLE condominio
  ADD COLUMN IF NOT EXISTS permite_soporte BOOLEAN NOT NULL DEFAULT false;
