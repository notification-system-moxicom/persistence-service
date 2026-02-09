-- +goose Up
-- +goose StatementBegin
CREATE TYPE notification_status AS ENUM ('pending', 'sent');

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    system_id UUID NOT NULL,
    content TEXT NOT NULL,
    status notification_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_notifications_system FOREIGN KEY (system_id) REFERENCES systems(id) ON DELETE CASCADE
);

CREATE TABLE notification_recipients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    notification_id UUID NOT NULL,
    user_id UUID NOT NULL,
    CONSTRAINT fk_recipients_notification FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
    CONSTRAINT fk_recipients_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_notifications_system_id ON notifications(system_id);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_recipients_notification_id ON notification_recipients(notification_id);
CREATE INDEX idx_recipients_user_id ON notification_recipients(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notification_recipients;
DROP TABLE IF EXISTS notifications;
DROP TYPE IF EXISTS notification_status;
-- +goose StatementEnd
