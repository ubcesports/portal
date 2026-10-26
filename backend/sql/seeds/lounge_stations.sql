BEGIN;

INSERT INTO lounge_stations (
    code,
    label,
    station_type,
    closed,
    grid_row,
    grid_column,
    width,
    height
)
VALUES
    -- Right wall PCs
    ('pc-01',    'PC 1',     'pc',       FALSE, 4, 4, 1, 1),
    ('pc-02',    'PC 2',     'pc',       FALSE, 5, 4, 1, 1),
    ('pc-03',    'PC 3',     'pc',       FALSE, 6, 4, 1, 1),

    -- Bottom row desks
    ('pc-04',    'PC 4',     'pc',       FALSE, 8, 1, 1, 1),
    ('pc-05',    'PC 5',     'pc',       FALSE, 8, 0, 1, 1),

    -- Middle pod (bottom side: 06, 07, 08)
    ('pc-06',    'PC 6',     'pc',       FALSE, 6, 0, 1, 1),
    ('pc-07',    'PC 7',     'pc',       FALSE, 6, 1, 1, 1),
    ('pc-08',    'PC 8',     'pc',       FALSE, 6, 2, 1, 1),

    -- Middle pod (top side: 11, 10, 09)
    ('pc-09',    'PC 9',     'pc',       FALSE, 5, 2, 1, 1),
    ('pc-10',    'PC 10',    'pc',       FALSE, 5, 1, 1, 1),
    ('pc-11',    'PC 11',    'pc',       FALSE, 5, 0, 1, 1),

    -- Top pod (bottom side: 12, 13)
    ('pc-12',    'PC 12',    'pc',       FALSE, 3, 0, 1, 1),
    ('pc-13',    'PC 13',    'pc',       FALSE, 3, 1, 1, 1),

    -- Top pod (PC 14 double-height desk spanning rows 2 & 3)
    ('pc-14',    'PC 14',    'pc',       FALSE, 2, 2, 1, 2),

    -- Top pod (top side: 16, 15)
    ('pc-15',    'PC 15',    'pc',       FALSE, 2, 1, 1, 1),
    ('pc-16',    'PC 16',    'pc',       FALSE, 2, 0, 1, 1),

    -- Top row desks (17, 18, 19)
    ('pc-17',    'PC 17',    'pc',       FALSE, 0, 0, 1, 1),
    ('pc-18',    'PC 18',    'pc',       FALSE, 0, 1, 1, 1),
    ('pc-19',    'PC 19',    'pc',       FALSE, 0, 2, 1, 1),

    -- Top right kiosk / Sign-in PC ("You")
    ('check-in', 'Check-in', 'check_in', FALSE, 0, 4, 1, 1),

    -- Arcades on bottom right wall ("C1", "C2")
    ('cab-01',   'C1',       'cab',      FALSE, 7, 4, 1, 1),
    ('cab-02',   'C2',       'cab',      FALSE, 8, 4, 1, 1)
ON CONFLICT (code) DO UPDATE SET
    label = EXCLUDED.label,
    station_type = EXCLUDED.station_type,
    closed = EXCLUDED.closed,
    grid_row = EXCLUDED.grid_row,
    grid_column = EXCLUDED.grid_column,
    width = EXCLUDED.width,
    height = EXCLUDED.height,
    updated_at = NOW();

COMMIT;
