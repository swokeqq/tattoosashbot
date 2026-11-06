-- +goose Up
CREATE TABLE conversations (
    connection_id TEXT        NOT NULL,
    chat_id       BIGINT      NOT NULL,
    state         TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (connection_id, chat_id)
);

-- +goose Down
DROP TABLE conversations;