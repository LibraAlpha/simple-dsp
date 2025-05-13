CREATE TABLE IF NOT EXISTS advertisers (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL,
    budget DECIMAL(20,4) NOT NULL,
    create_time TIMESTAMP NOT NULL,
    update_time TIMESTAMP NOT NULL
);

CREATE INDEX idx_advertisers_status ON advertisers(status);
CREATE INDEX idx_advertisers_create_time ON advertisers(create_time); 