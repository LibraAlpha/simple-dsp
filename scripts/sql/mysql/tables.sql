-- 创建数据库
CREATE DATABASE IF NOT EXISTS simple_dsp DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE simple_dsp;

-- 素材表
CREATE TABLE IF NOT EXISTS creatives (
    id VARCHAR(32) NOT NULL COMMENT '素材ID',
    name VARCHAR(255) NOT NULL COMMENT '素材名称',
    type VARCHAR(32) NOT NULL COMMENT '素材类型(image/video/html)',
    format VARCHAR(32) NOT NULL COMMENT '文件格式',
    size BIGINT NOT NULL COMMENT '文件大小',
    width INT COMMENT '宽度',
    height INT COMMENT '高度',
    duration DECIMAL(10,2) COMMENT '视频时长',
    url VARCHAR(1024) NOT NULL COMMENT '访问URL',
    storage_path VARCHAR(1024) NOT NULL COMMENT '存储路径',
    status VARCHAR(32) NOT NULL COMMENT '状态(active/inactive/deleted)',
    create_time DATETIME NOT NULL COMMENT '创建时间',
    update_time DATETIME NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    INDEX idx_status (status),
    INDEX idx_type (type),
    INDEX idx_create_time (create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='素材表';

-- 素材标签表
CREATE TABLE IF NOT EXISTS creative_tags (
    creative_id VARCHAR(32) NOT NULL COMMENT '素材ID',
    tag VARCHAR(64) NOT NULL COMMENT '标签',
    create_time DATETIME NOT NULL COMMENT '创建时间',
    PRIMARY KEY (creative_id, tag),
    INDEX idx_tag (tag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='素材标签表';

-- 素材组表
CREATE TABLE IF NOT EXISTS creative_groups (
    id VARCHAR(32) NOT NULL COMMENT '素材组ID',
    name VARCHAR(255) NOT NULL COMMENT '素材组名称',
    description TEXT COMMENT '描述',
    status VARCHAR(32) NOT NULL COMMENT '状态(active/inactive/deleted)',
    create_time DATETIME NOT NULL COMMENT '创建时间',
    update_time DATETIME NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='素材组表';

-- 素材组关联表
CREATE TABLE IF NOT EXISTS creative_group_relations (
    group_id VARCHAR(32) NOT NULL COMMENT '素材组ID',
    creative_id VARCHAR(32) NOT NULL COMMENT '素材ID',
    create_time DATETIME NOT NULL COMMENT '创建时间',
    PRIMARY KEY (group_id, creative_id),
    INDEX idx_creative_id (creative_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='素材组关联表';

-- 素材审核记录表
CREATE TABLE IF NOT EXISTS creative_audit_records (
    id VARCHAR(32) NOT NULL COMMENT '审核记录ID',
    creative_id VARCHAR(32) NOT NULL COMMENT '素材ID',
    status VARCHAR(32) NOT NULL COMMENT '审核状态(pending/approved/rejected/revision)',
    reviewer VARCHAR(64) COMMENT '审核人',
    comments TEXT COMMENT '审核意见',
    create_time DATETIME NOT NULL COMMENT '创建时间',
    update_time DATETIME NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    INDEX idx_creative_id (creative_id),
    INDEX idx_status (status),
    INDEX idx_create_time (create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='素材审核记录表';

-- 素材版本表
CREATE TABLE IF NOT EXISTS creative_versions (
    id VARCHAR(32) NOT NULL COMMENT '版本ID',
    creative_id VARCHAR(32) NOT NULL COMMENT '素材ID',
    version INT NOT NULL COMMENT '版本号',
    changes TEXT COMMENT '变更说明',
    storage_path VARCHAR(1024) NOT NULL COMMENT '存储路径',
    status VARCHAR(32) NOT NULL COMMENT '状态(active/inactive)',
    creator VARCHAR(64) NOT NULL COMMENT '创建人',
    create_time DATETIME NOT NULL COMMENT '创建时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_creative_version (creative_id, version),
    INDEX idx_creative_id (creative_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='素材版本表';

-- 广告表
CREATE TABLE IF NOT EXISTS ads (
    id VARCHAR(32) NOT NULL COMMENT '广告ID',
    title VARCHAR(255) NOT NULL COMMENT '标题',
    description TEXT COMMENT '描述',
    image_url VARCHAR(1024) COMMENT '图片URL',
    landing_url VARCHAR(1024) NOT NULL COMMENT '落地页URL',
    width INT NOT NULL COMMENT '宽度',
    height INT NOT NULL COMMENT '高度',
    budget_id VARCHAR(32) NOT NULL COMMENT '预算ID',
    status VARCHAR(32) NOT NULL COMMENT '状态(active/inactive/deleted)',
    create_time DATETIME NOT NULL COMMENT '创建时间',
    update_time DATETIME NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    INDEX idx_status (status),
    INDEX idx_budget_id (budget_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='广告表';

-- 竞价记录表
CREATE TABLE IF NOT EXISTS bid_records (
    id VARCHAR(32) NOT NULL COMMENT '记录ID',
    request_id VARCHAR(64) NOT NULL COMMENT '请求ID',
    user_id VARCHAR(64) NOT NULL COMMENT '用户ID',
    device_id VARCHAR(64) COMMENT '设备ID',
    ip VARCHAR(64) COMMENT 'IP地址',
    slot_id VARCHAR(64) NOT NULL COMMENT '广告位ID',
    ad_id VARCHAR(32) COMMENT '广告ID',
    bid_price DECIMAL(12,6) COMMENT '出价',
    win_price DECIMAL(12,6) COMMENT '获胜价格',
    status VARCHAR(32) NOT NULL COMMENT '状态(success/fail)',
    error_msg TEXT COMMENT '错误信息',
    create_time DATETIME NOT NULL COMMENT '创建时间',
    PRIMARY KEY (id),
    INDEX idx_request_id (request_id),
    INDEX idx_user_id (user_id),
    INDEX idx_ad_id (ad_id),
    INDEX idx_create_time (create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='竞价记录表'; 