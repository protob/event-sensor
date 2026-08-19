-- +goose Up

-- Seed default admin user (username: admin, password: password).
-- bcrypt hash generated with bcrypt.DefaultCost (cost 10).
-- Uses INSERT OR IGNORE so existing 'admin' usernames are not overwritten.
INSERT OR IGNORE INTO users (id, username, email, password, role, active)
VALUES (
    '63bae288-3c8e-47d6-9ac9-ad994166a3b3',
    'admin',
    'admin@app.localdev',
    '$2a$10$YIjhIFyPj0l//d523Jgni.SbjAdT7CMiNy0oy9YyaGs3WbI16RiJy',
    'admin',
    1
);

-- +goose Down
DELETE FROM users WHERE username = 'admin';
