BEGIN;

INSERT INTO lounge_stations (
    id,
    label,
    station_type,
    closed,
    "row",
    col,
    width,
    height,
    active_session_id
)
VALUES
    ('c1000000-0000-4000-8000-000000000001', 'PC 1',  'pc',         FALSE, 4, 4, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000002', 'PC 2',  'pc',         FALSE, 5, 4, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000003', 'PC 3',  'pc',         FALSE, 6, 4, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000004', 'PC 4',  'pc',         FALSE, 8, 1, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000005', 'PC 5',  'pc',         FALSE, 8, 0, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000006', 'PC 6',  'pc',         FALSE, 6, 0, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000007', 'PC 7',  'pc',         FALSE, 6, 1, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000008', 'PC 8',  'pc',         FALSE, 6, 2, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000009', 'PC 9',  'pc',         FALSE, 5, 2, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000010', 'PC 10', 'pc',         FALSE, 5, 1, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000011', 'PC 11', 'pc',         FALSE, 5, 0, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000012', 'PC 12', 'pc',         FALSE, 3, 0, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000013', 'PC 13', 'pc',         FALSE, 3, 1, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000014', 'PC 14', 'pc',         FALSE, 2, 2, 1, 2, NULL),
    ('c1000000-0000-4000-8000-000000000015', 'PC 15', 'pc',         FALSE, 2, 1, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000016', 'PC 16', 'pc',         FALSE, 2, 0, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000017', 'PC 17', 'pc',         FALSE, 0, 0, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000018', 'PC 18', 'pc',         FALSE, 0, 1, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000019', 'PC 19', 'pc',         FALSE, 0, 2, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000020', 'PC 20', 'Sign in pc', FALSE, 0, 4, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000021', 'C1',    'arcade',     FALSE, 7, 4, 1, 1, NULL),
    ('c1000000-0000-4000-8000-000000000022', 'C2',    'arcade',     FALSE, 8, 4, 1, 1, NULL)
ON CONFLICT (id) DO UPDATE SET
    label = EXCLUDED.label,
    station_type = EXCLUDED.station_type,
    closed = EXCLUDED.closed,
    "row" = EXCLUDED."row",
    col = EXCLUDED.col,
    width = EXCLUDED.width,
    height = EXCLUDED.height,
    active_session_id = EXCLUDED.active_session_id;

COMMIT;
