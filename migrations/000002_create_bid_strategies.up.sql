CREATE TABLE bid_strategies (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL COMMENT '策略名称',
    bid_type ENUM('CPC', 'CPM') NOT NULL COMMENT '计费类型',
    price DECIMAL(10,4) NOT NULL COMMENT '出价，CPC单位为元，CPM单位为分',
    daily_budget DECIMAL(10,2) NOT NULL COMMENT '日预算，单位为元',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：0-禁用，1-启用',
    is_price_locked TINYINT NOT NULL DEFAULT 1 COMMENT '出价是否锁定：0-未锁定，1-锁定',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_bid_type (bid_type),
    INDEX idx_status (status),
    INDEX idx_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='出价策略表';

CREATE TABLE bid_strategy_creatives (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    strategy_id BIGINT NOT NULL COMMENT '策略ID',
    creative_id BIGINT NOT NULL COMMENT '素材ID',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：0-禁用，1-启用',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_strategy_id (strategy_id),
    INDEX idx_creative_id (creative_id),
    INDEX idx_status (status),
    FOREIGN KEY (strategy_id) REFERENCES bid_strategies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='出价策略素材关联表'; 