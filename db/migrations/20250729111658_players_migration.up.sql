CREATE TABLE players (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    height DECIMAL(5,2) NOT NULL,
    weight DECIMAL(5,2) NOT NULL,
    position VARCHAR(20) CHECK (position IN ('penyerang', 'gelandang', 'bertahan', 'penjaga_gawang')),
    jersey_number INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE UNIQUE INDEX unique_team_jersey_number_not_deleted
ON players(team_id, jersey_number)
WHERE deleted_at IS NULL;
