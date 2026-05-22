CREATE TABLE IF NOT EXISTS guests (
  id               BIGSERIAL PRIMARY KEY,
  email            TEXT        NOT NULL UNIQUE,   -- stored lower-cased + trimmed
  first_name       TEXT        NOT NULL,
  last_name        TEXT        NOT NULL,
  plus_one_allowed BOOLEAN     NOT NULL DEFAULT false,
  plus_one_name    TEXT,                          -- known +1 name from the list, used as a default
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE rsvps ADD COLUMN IF NOT EXISTS guest_id             BIGINT REFERENCES guests(id);
ALTER TABLE rsvps ADD COLUMN IF NOT EXISTS plus_one_attending   BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE rsvps ADD COLUMN IF NOT EXISTS plus_one_name        TEXT;
ALTER TABLE rsvps ADD COLUMN IF NOT EXISTS plus_one_meal_choice TEXT;

-- One RSVP per guest, so a returning guest's submission updates rather than duplicates.
CREATE UNIQUE INDEX IF NOT EXISTS rsvps_guest_id_key ON rsvps (guest_id);
