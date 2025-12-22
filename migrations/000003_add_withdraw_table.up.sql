CREATE TABLE withdraws (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    order_id VARCHAR(255) NOT NULL,
    sum INTEGER NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_withdraws_user_id ON withdraws(user_id);
