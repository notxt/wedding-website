CREATE TABLE IF NOT EXISTS rsvps (
  id              BIGSERIAL PRIMARY KEY,
  submitted_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  attending       BOOLEAN     NOT NULL,
  party_names     TEXT        NOT NULL,
  meal_choice     TEXT,
  dairy_allergy   BOOLEAN     NOT NULL DEFAULT false,
  gluten_allergy  BOOLEAN     NOT NULL DEFAULT false,
  other_allergies TEXT,
  remote_addr     TEXT
);
