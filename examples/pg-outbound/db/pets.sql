CREATE TABLE pets (
  id INT PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  prey VARCHAR(100),
  is_finicky BOOL NOT NULL,
  timestamp TIMESTAMP
);
INSERT INTO pets VALUES (1, 'Splodge', NULL, false, '2026-04-20 12:30:00');
INSERT INTO pets VALUES (2, 'Kiki', 'Cicadas', false, '2026-04-20 12:30:00');
INSERT INTO pets VALUES (3, 'Slats', 'Temptations', true, '2026-04-20 12:30:00');
