-- SQLite 多页功能迁移脚本

-- 创建页面表
CREATE TABLE IF NOT EXISTS "item_pages" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "title" TEXT DEFAULT '页面',
    "icon" TEXT DEFAULT 'material-symbols:home-outline',
    "sort" INTEGER DEFAULT 9999,
    "user_id" INTEGER NOT NULL,
    "created_at" DATETIME DEFAULT CURRENT_TIMESTAMP,
    "updated_at" DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS "idx_user_sort" ON "item_pages" ("user_id", "sort");

-- 为分组表添加page_id字段
ALTER TABLE "item_icon_groups" ADD COLUMN "page_id" INTEGER;
CREATE INDEX IF NOT EXISTS "idx_page_id" ON "item_icon_groups" ("page_id");

-- 数据迁移：为每个用户创建默认页面
INSERT INTO "item_pages" ("user_id", "title", "icon", "sort")
SELECT DISTINCT "user_id", '首页', 'material-symbols:home-outline', 1
FROM "item_icon_groups"
WHERE "page_id" IS NULL
GROUP BY "user_id";

-- 将现有分组关联到对应的默认页面
UPDATE "item_icon_groups"
SET "page_id" = (
    SELECT "id" FROM "item_pages" 
    WHERE "item_pages"."user_id" = "item_icon_groups"."user_id" 
    AND "item_pages"."sort" = 1
    LIMIT 1
)
WHERE "page_id" IS NULL;
