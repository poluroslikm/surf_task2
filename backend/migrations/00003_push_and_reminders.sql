-- +goose Up
-- push_subscriptions: one row per browser/device Web Push registration (FR-21/FR-22, FEAT-02).
-- endpoint is globally unique and is the upsert key (FEAT-03) — a subscription's owner
-- (client_id) can legitimately change if the same browser endpoint gets reused by a different
-- logged-in client on a shared device.
CREATE TABLE push_subscriptions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id uuid NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    endpoint text NOT NULL UNIQUE,
    p256dh text NOT NULL,
    auth text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX push_subscriptions_client_id_idx ON push_subscriptions (client_id);

-- booking_reminders: guards the 24h-reminder worker (FEAT-04, FR-21) against re-sending on
-- worker restart/re-tick — a row's presence means a reminder has already been attempted for
-- that booking, regardless of push delivery success.
--
-- Deliberate deviation from the original sketch (`ALTER TABLE bookings ADD COLUMN
-- reminder_sent_at`): verified live against this environment's actual DATABASE_URL role
-- (chef_stol) — `bookings` is owned by a different role (`postgres`), and chef_stol, while
-- granted full DML (SELECT/INSERT/UPDATE/DELETE) on it, has no ALTER privilege and there is no
-- reachable superuser session to grant one (no known password, no permission to reload/restart
-- the Postgres service in this environment). A brand-new table is owned by whoever creates it
-- (chef_stol), sidestepping the ownership requirement entirely while providing the exact same
-- "has a reminder already been attempted for this booking" idempotency check.
CREATE TABLE booking_reminders (
    booking_id uuid PRIMARY KEY REFERENCES bookings(id) ON DELETE CASCADE,
    sent_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE booking_reminders;
DROP TABLE push_subscriptions;
