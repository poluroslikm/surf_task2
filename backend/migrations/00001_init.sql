-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE clients (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT clients_email_format_chk CHECK (email ~ '^[^@\s]+@[^@\s]+\.[^@\s]+$')
);

CREATE TABLE auth_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id uuid NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    token_hash text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    CONSTRAINT auth_sessions_expiry_chk CHECK (expires_at > created_at)
);

-- Program/Chef: read-only справочники для клиента (02-domain.md → «Границы скоупа»).
-- Нет capacity_cap на программе — вместимость не привязана формулой к сложности
-- (в отличие от primer/-референса, где routes.capacity_cap жёстко зависит от routes.type).
CREATE TABLE programs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    difficulty text NOT NULL,
    photo_url text NOT NULL,
    ingredients text[] NOT NULL DEFAULT '{}',
    allergens text[] NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT programs_difficulty_chk CHECK (difficulty IN ('novice', 'experienced'))
);

CREATE TABLE chefs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE slots (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    program_id uuid NOT NULL REFERENCES programs(id) ON DELETE RESTRICT,
    chef_id uuid NOT NULL REFERENCES chefs(id) ON DELETE RESTRICT,
    start_at timestamptz NOT NULL,
    total_seats integer NOT NULL,
    free_seats integer NOT NULL,
    price integer NOT NULL,
    status text NOT NULL DEFAULT 'scheduled',
    cancellation_reason text,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT slots_status_chk CHECK (status IN ('scheduled', 'cancelled')),
    CONSTRAINT slots_seats_chk CHECK (total_seats > 0 AND free_seats >= 0 AND free_seats <= total_seats),
    CONSTRAINT slots_price_chk CHECK (price >= 0),
    CONSTRAINT slots_cancellation_reason_chk CHECK (
        (status = 'cancelled' AND cancellation_reason IS NOT NULL)
        OR (status = 'scheduled' AND cancellation_reason IS NULL)
    )
);

-- Bookings: схема на будущее (BE-06+ ещё не реализован в этой итерации). Без seats_count
-- (одна запись — одно место) и без счётчиков прокатного фонда (FR-9 — фонд не ограничен).
CREATE TABLE bookings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slot_id uuid NOT NULL REFERENCES slots(id) ON DELETE RESTRICT,
    client_id uuid NOT NULL REFERENCES clients(id) ON DELETE RESTRICT,
    equipment_choice text NOT NULL,
    status text NOT NULL DEFAULT 'active',
    price_total integer NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    cancelled_at timestamptz,
    rating_stars smallint,
    rating_comment text,
    rating_created_at timestamptz,
    CONSTRAINT bookings_equipment_choice_chk CHECK (equipment_choice IN ('own', 'rental')),
    CONSTRAINT bookings_status_chk CHECK (status IN ('active', 'cancelled', 'late_cancel', 'cancelled_by_studio')),
    CONSTRAINT bookings_price_chk CHECK (price_total >= 0),
    CONSTRAINT bookings_cancelled_at_chk CHECK (
        (status = 'active' AND cancelled_at IS NULL)
        OR (status <> 'active' AND cancelled_at IS NOT NULL)
    ),
    CONSTRAINT bookings_rating_stars_chk CHECK (rating_stars IS NULL OR rating_stars BETWEEN 1 AND 5),
    CONSTRAINT bookings_rating_consistency_chk CHECK (
        (rating_stars IS NULL AND rating_created_at IS NULL)
        OR (rating_stars IS NOT NULL AND rating_created_at IS NOT NULL)
    )
);

CREATE UNIQUE INDEX bookings_active_client_slot_uidx ON bookings (client_id, slot_id) WHERE status = 'active';
CREATE INDEX auth_sessions_client_id_idx ON auth_sessions (client_id);
CREATE INDEX auth_sessions_token_hash_idx ON auth_sessions (token_hash);
CREATE INDEX slots_start_at_idx ON slots (start_at);
CREATE INDEX slots_status_idx ON slots (status);
CREATE INDEX slots_program_id_idx ON slots (program_id);
CREATE INDEX slots_chef_id_idx ON slots (chef_id);
CREATE INDEX bookings_slot_id_idx ON bookings (slot_id);
CREATE INDEX bookings_client_id_idx ON bookings (client_id);
CREATE INDEX bookings_status_idx ON bookings (status);

-- +goose Down
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS slots;
DROP TABLE IF EXISTS chefs;
DROP TABLE IF EXISTS programs;
DROP TABLE IF EXISTS auth_sessions;
DROP TABLE IF EXISTS clients;
