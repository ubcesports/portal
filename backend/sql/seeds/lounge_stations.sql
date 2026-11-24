BEGIN;

INSERT INTO lounge_stations (
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
    ('PC 1',     'pc',       FALSE, 4, 4, 1, 1),
    ('PC 2',     'pc',       FALSE, 5, 4, 1, 1),
    ('PC 3',     'pc',       FALSE, 6, 4, 1, 1),

    -- Bottom row desks
    ('PC 4',     'pc',       FALSE, 8, 1, 1, 1),
    ('PC 5',     'pc',       FALSE, 8, 0, 1, 1),

    -- Middle pod (bottom side: 06, 07, 08)
    ('PC 6',     'pc',       FALSE, 6, 0, 1, 1),
    ('PC 7',     'pc',       FALSE, 6, 1, 1, 1),
    ('PC 8',     'pc',       FALSE, 6, 2, 1, 1),

    -- Middle pod (top side: 11, 10, 09)
    ('PC 9',     'pc',       FALSE, 5, 2, 1, 1),
    ('PC 10',    'pc',       FALSE, 5, 1, 1, 1),
    ('PC 11',    'pc',       FALSE, 5, 0, 1, 1),

    -- Top pod (bottom side: 12, 13)
    ('PC 12',    'pc',       FALSE, 3, 0, 1, 1),
    ('PC 13',    'pc',       FALSE, 3, 1, 1, 1),

    -- Top pod (PC 14 double-height desk spanning rows 2 & 3)
    ('PC 14',    'pc',       FALSE, 2, 2, 1, 2),

    -- Top pod (top side: 16, 15)
    ('PC 15',    'pc',       FALSE, 2, 1, 1, 1),
    ('PC 16',    'pc',       FALSE, 2, 0, 1, 1),

    -- Top row desks (17, 18, 19)
    ('PC 17',    'pc',       FALSE, 0, 0, 1, 1),
    ('PC 18',    'pc',       FALSE, 0, 1, 1, 1),
    ('PC 19',    'pc',       FALSE, 0, 2, 1, 1),

    -- Top right kiosk / Sign-in PC ("You")
    ('Check-in', 'check_in', FALSE, 0, 4, 1, 1),

    -- Arcades on bottom right wall ("C1", "C2")
    ('C1',       'cab',      FALSE, 7, 4, 1, 1),
    ('C2',       'cab',      FALSE, 8, 4, 1, 1);

COMMIT;
