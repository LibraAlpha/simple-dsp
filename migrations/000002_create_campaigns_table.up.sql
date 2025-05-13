CREATE TABLE IF NOT EXISTS campaigns (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    advertiser_id VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    budget DECIMAL(20,4) NOT NULL,
    bid_strategy VARCHAR(32) NOT NULL,
    targeting JSONB,
    tracking_configs JSONB,
    update_time TIMESTAMP NOT NULL,
    create_time TIMESTAMP NOT NULL,
    
    CONSTRAINT fk_campaigns_advertiser FOREIGN KEY (advertiser_id)
        REFERENCES advertisers (id) ON DELETE CASCADE
);

CREATE INDEX idx_campaigns_advertiser ON campaigns(advertiser_id);
CREATE INDEX idx_campaigns_status ON campaigns(status);
CREATE INDEX idx_campaigns_time ON campaigns(start_time, end_time); 