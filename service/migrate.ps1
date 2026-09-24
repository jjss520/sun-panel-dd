# 多页功能 - 数据库迁移脚本
Write-Host "开始执行数据库迁移..." -ForegroundColor Green

# 读取SQL文件
$sqlFile = "migrate_add_pages.sql"
if (Test-Path $sqlFile) {
    Write-Host "找到迁移文件: $sqlFile" -ForegroundColor Yellow
    
    # 提示用户手动执行
    Write-Host "`n请按照以下步骤执行迁移:" -ForegroundColor Cyan
    Write-Host "1. 连接到MySQL数据库" -ForegroundColor White
    Write-Host "2. 选择sun_panel数据库" -ForegroundColor White  
    Write-Host "3. 执行以下SQL文件: $sqlFile" -ForegroundColor White
    Write-Host "`n或者在MySQL命令行中执行:" -ForegroundColor Cyan
    Write-Host "source $sqlFile" -ForegroundColor Yellow
    Write-Host "\. $sqlFile" -ForegroundColor Yellow
    
    Write-Host "`n迁移完成后,重启后端服务即可生效!" -ForegroundColor Green
} else {
    Write-Host "错误: 找不到迁移文件 $sqlFile" -ForegroundColor Red
}
