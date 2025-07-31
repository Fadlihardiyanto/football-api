CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

CREATE INDEX idx_teams_name ON teams(name);
CREATE INDEX idx_teams_deleted_at ON teams(deleted_at);

CREATE INDEX idx_players_team_id ON players(team_id);
CREATE INDEX idx_players_position ON players(position);
CREATE INDEX idx_players_deleted_at ON players(deleted_at);

CREATE INDEX idx_matches_home_team_id ON matches(home_team_id);
CREATE INDEX idx_matches_away_team_id ON matches(away_team_id);
CREATE INDEX idx_matches_status ON matches(status);
CREATE INDEX idx_matches_match_date ON matches(match_date);
CREATE INDEX idx_matches_deleted_at ON matches(deleted_at);

CREATE INDEX idx_goals_match_id ON goals(match_id);
CREATE INDEX idx_goals_player_id ON goals(player_id);
CREATE INDEX idx_goals_deleted_at ON goals(deleted_at);
