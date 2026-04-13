#!/usr/bin/env pwsh

# Docker 镜像构建和推送脚本
# 使用方法：.\build-and-push.ps1 -Username "your-username" -Registry "ghcr.io"

param(
    [Parameter(Mandatory=$true)]
    [string]$Username,
    
    [Parameter(Mandatory=$false)]
    [ValidateSet("ghcr.io", "docker.io", "registry.cn-hangzhou.aliyuncs.com")]
    [string]$Registry = "ghcr.io",
    
    [Parameter(Mandatory=$false)]
    [string]$Version = "latest"
)

$ImageName = "sun-panel-dd"
$WorkDir = $PSScriptRoot

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Docker 镜像构建和推送" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 设置镜像名称
if ($Registry -eq "docker.io") {
    $FullImageName = "$Username/$ImageName"
} else {
    $FullImageName = "$Registry/$Username/$ImageName"
}

Write-Host "📦 镜像名称：$FullImageName:$Version" -ForegroundColor Green
Write-Host "📁 工作目录：$WorkDir" -ForegroundColor Green
Write-Host ""

# 步骤 1: 构建镜像
Write-Host "🔨 步骤 1/4: 构建 Docker 镜像..." -ForegroundColor Yellow
Set-Location $WorkDir
docker build -t $ImageName:latest .

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 镜像构建失败！" -ForegroundColor Red
    exit 1
}
Write-Host "✅ 镜像构建成功！" -ForegroundColor Green
Write-Host ""

# 步骤 2: 标记镜像
Write-Host "🏷️  步骤 2/4: 标记镜像..." -ForegroundColor Yellow
docker tag $ImageName:latest $FullImageName:$Version

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 标记镜像失败！" -ForegroundColor Red
    exit 1
}
Write-Host "✅ 镜像标记成功：$FullImageName:$Version" -ForegroundColor Green
Write-Host ""

# 步骤 3: 登录 Docker 仓库
Write-Host "🔐 步骤 3/4: 登录 Docker 仓库 ($Registry)..." -ForegroundColor Yellow
docker login $Registry

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 登录失败！请检查账号密码。" -ForegroundColor Red
    exit 1
}
Write-Host "✅ 登录成功！" -ForegroundColor Green
Write-Host ""

# 步骤 4: 推送镜像
Write-Host "🚀 步骤 4/4: 推送镜像到 $Registry..." -ForegroundColor Yellow
docker push $FullImageName:$Version

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 推送失败！" -ForegroundColor Red
    exit 1
}
Write-Host "✅ 推送成功！" -ForegroundColor Green
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  🎉 完成！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "镜像信息:" -ForegroundColor Cyan
Write-Host "  仓库：$FullImageName" -ForegroundColor White
Write-Host "  版本：$Version" -ForegroundColor White
Write-Host "  本地标签：$ImageName:latest" -ForegroundColor White
Write-Host ""
Write-Host "使用示例:" -ForegroundColor Cyan
Write-Host "  docker run -d -p 3002:3002 $FullImageName:$Version" -ForegroundColor Gray
Write-Host ""
