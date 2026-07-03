-- +goose Up
INSERT INTO routes (id, name, type, capacity_cap, duration_min, geometry)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'Острова и каналы', 'novice', 8, 90, '[[59.978,30.262],[59.981,30.271],[59.976,30.285]]'),
    ('22222222-2222-2222-2222-222222222222', 'Большая вода', 'experienced', 12, 120, '[[59.942,30.226],[59.951,30.214],[59.963,30.232]]')
ON CONFLICT (id) DO NOTHING;

INSERT INTO instructors (id, name)
VALUES
    ('33333333-3333-3333-3333-333333333333', 'Мария'),
    ('44444444-4444-4444-4444-444444444444', 'Алексей')
ON CONFLICT (id) DO NOTHING;

INSERT INTO slots (
    id,
    route_id,
    instructor_id,
    start_at,
    total_seats,
    free_seats,
    rental_boards_total,
    free_rental_boards,
    price,
    rental_price,
    meeting_point,
    meeting_point_lat,
    meeting_point_lng,
    status
)
VALUES
    (
        '55555555-5555-5555-5555-555555555555',
        '11111111-1111-1111-1111-111111111111',
        '33333333-3333-3333-3333-333333333333',
        '2030-06-01 10:00:00+03',
        8,
        8,
        12,
        12,
        2500,
        800,
        'Лодочная станция у Елагина моста',
        59.978,
        30.262,
        'scheduled'
    ),
    (
        '66666666-6666-6666-6666-666666666666',
        '22222222-2222-2222-2222-222222222222',
        '44444444-4444-4444-4444-444444444444',
        '2030-06-02 12:00:00+03',
        12,
        12,
        12,
        10,
        3200,
        800,
        'Пирс у Крестовского острова',
        59.942,
        30.226,
        'scheduled'
    )
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DELETE FROM slots WHERE id IN ('55555555-5555-5555-5555-555555555555', '66666666-6666-6666-6666-666666666666');
DELETE FROM instructors WHERE id IN ('33333333-3333-3333-3333-333333333333', '44444444-4444-4444-4444-444444444444');
DELETE FROM routes WHERE id IN ('11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222');
