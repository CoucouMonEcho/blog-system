-- 创建数据库
CREATE DATABASE IF NOT EXISTS blog_system CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE blog_system;

-- 用户表
CREATE TABLE IF NOT EXISTS blog_user
(
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    username   VARCHAR(50)  NOT NULL UNIQUE,
    email      VARCHAR(100) NOT NULL UNIQUE,
    password   VARCHAR(255) NOT NULL,
    role       VARCHAR(20)  NOT NULL DEFAULT 'user' COMMENT 'user/admin',
    avatar     VARCHAR(255),
    status     TINYINT      NOT NULL DEFAULT 0 COMMENT '0: 正常, 1: 禁用',
    created_at TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP             DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_status (status)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

-- 单级分类表
CREATE TABLE IF NOT EXISTS blog_category
(
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    sort        INT       DEFAULT 0,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_slug (slug),
    INDEX idx_sort (sort)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

-- 标签表
CREATE TABLE IF NOT EXISTS blog_tag
(
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    name       VARCHAR(50) NOT NULL UNIQUE,
    slug       VARCHAR(50) NOT NULL UNIQUE,
    color      VARCHAR(20) DEFAULT '#666666',
    created_at TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP   DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_slug (slug)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

-- 文章表
CREATE TABLE IF NOT EXISTS blog_article
(
    id            BIGINT AUTO_INCREMENT PRIMARY KEY,
    title         VARCHAR(200) NOT NULL,
    slug          VARCHAR(200) NOT NULL UNIQUE,
    content       LONGTEXT     NOT NULL,
    summary       TEXT,
    cover         VARCHAR(500),
    author_id     BIGINT       NOT NULL,
    category_id   BIGINT       NOT NULL,
    status        TINYINT      NOT NULL DEFAULT 0 COMMENT '0: 草稿, 1: 发布, 2: 私密',
    view_count    BIGINT                DEFAULT 0,
    like_count    BIGINT                DEFAULT 0,
    is_top        BOOLEAN               DEFAULT FALSE,
    is_recommend  BOOLEAN               DEFAULT FALSE,
    meta_title    VARCHAR(200),
    meta_desc     VARCHAR(500),
    meta_keywords VARCHAR(200),
    published_at  TIMESTAMP    NULL,
    created_at    TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP             DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_slug (slug),
    INDEX idx_author_id (author_id),
    INDEX idx_category_id (category_id),
    INDEX idx_status (status),
    INDEX idx_published_at (published_at),
    FOREIGN KEY (author_id) REFERENCES blog_user (id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES blog_category (id) ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

-- 文章标签关联表
CREATE TABLE IF NOT EXISTS blog_article_tags
(
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    article_id BIGINT NOT NULL,
    tag_id     BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_article_tag (article_id, tag_id),
    INDEX idx_article_id (article_id),
    INDEX idx_tag_id (tag_id),
    FOREIGN KEY (article_id) REFERENCES blog_article (id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES blog_tag (id) ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

-- 统计表
CREATE TABLE IF NOT EXISTS blog_stat
(
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    type        VARCHAR(20) NOT NULL COMMENT '统计类型: view, like, favorite',
    target_id   BIGINT      NOT NULL COMMENT '目标ID',
    target_type VARCHAR(20) NOT NULL COMMENT '目标类型: article, user',
    user_id     BIGINT      NULL COMMENT '操作用户ID，可为空',
    count       BIGINT    DEFAULT 1,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_type_target (type, target_id, target_type, user_id),
    INDEX idx_type (type),
    INDEX idx_target (target_id, target_type),
    INDEX idx_user_id (user_id)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

-- 默认数据
INSERT INTO blog_user (username, email, password, role, status)
VALUES ('admin', 'admin@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', 0);

INSERT INTO blog_category (name, slug, description, sort)
VALUES ('技术', 'tech', '技术相关文章', 1),
       ('生活', 'life', '生活随笔', 2),
       ('随笔', 'essay', '个人随笔', 3);

INSERT INTO blog_tag (name, slug, color)
VALUES ('Go', 'go', '#00ADD8'),
       ('微服务', 'microservice', '#FF6B6B'),
       ('DDD', 'ddd', '#4ECDC4'),
       ('Docker', 'docker', '#45B7D1'),
       ('MySQL', 'mysql', '#F7DC6F');
