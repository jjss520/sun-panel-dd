# 执行SQLite数据库迁移
Write-Host "开始执行SQLite数据库迁移..." -ForegroundColor Green

# 找到容器中的数据库文件路径
$dbPath = "/data/database/database.db"
$sqlFile = "migrate_add_pages_sqlite.sql"

Write-Host "数据库路径: $dbPath" -ForegroundColor Yellow
Write-Host "SQL文件: $sqlFile" -ForegroundColor Yellow

# 复制SQL文件到容器
Write-Host "`n启动容器..." -ForegroundColor Cyan
docker start sun-panel-pages-test
Start-Sleep -Seconds 2

Write-Host "复制SQL文件到容器..." -ForegroundColor Cyan
docker cp $sqlFile sun-panel-pages-test:/tmp/migrate.sql

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ SQL文件复制成功" -ForegroundColor Green
    
    # 在容器中执行SQL
    Write-Host "`n执行SQL迁移..." -ForegroundColor Cyan
    docker exec sun-panel-pages-test sh -c "sqlite3 $dbPath < /tmp/migrate.sql"
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ 数据库迁移成功!" -ForegroundColor Green
        Write-Host "`n现在可以启动容器了:" -ForegroundColor Yellow
        Write-Host "docker start sun-panel-pages-test" -ForegroundColor White
    } else {
        Write-Host "✗ SQL执行失败" -ForegroundColor Red
    }
    
    # 清理临时文件
    docker exec sun-panel-pages-test rm -f /tmp/migrate.sql
} else {
    Write-Host "✗ 文件复制失败" -ForegroundColor Red
}
