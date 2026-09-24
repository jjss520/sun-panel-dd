-- 创建页面表
CREATE TABLE IF NOT EXISTS `item_pages` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `title` VARCHAR(100) DEFAULT '页面' COMMENT '页面标题',
    `icon` VARCHAR(50) DEFAULT 'material-symbols:home-outline' COMMENT '页面图标',
    `sort` INT DEFAULT 9999 COMMENT '排序',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX `idx_user_sort` (`user_id`, `sort`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='图标页面表';

-- 为分组表添加page_id字段
ALTER TABLE `item_icon_groups` 
ADD COLUMN `page_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '所属页面ID' AFTER `sort`,
ADD INDEX `idx_page_id` (`page_id`);

-- 数据迁移：为每个用户创建默认页面，并将现有分组关联到默认页面
INSERT INTO `item_pages` (`user_id`, `title`, `icon`, `sort`)
SELECT DISTINCT `user_id`, '首页', 'material-symbols:home-outline', 1
FROM `item_icon_groups`
WHERE `page_id` IS NULL
GROUP BY `user_id`;

-- 将现有分组关联到对应的默认页面
UPDATE `item_icon_groups` ig
INNER JOIN `item_pages` p ON ig.`user_id` = p.`user_id` AND p.`sort` = 1
SET ig.`page_id` = p.`id`
WHERE ig.`page_id` IS NULL;
