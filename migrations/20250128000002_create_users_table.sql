-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    system_id UUID NOT NULL,
    id_at_system VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(255),
    telegram_chat_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_users_system FOREIGN KEY (system_id) REFERENCES systems(id) ON DELETE CASCADE
);

CREATE INDEX idx_users_system_id ON users(system_id);
CREATE INDEX idx_users_id_at_system ON users(id_at_system);
CREATE UNIQUE INDEX idx_users_system_id_at_system ON users(system_id, id_at_system);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
