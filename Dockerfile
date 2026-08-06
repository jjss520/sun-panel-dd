# build frontend
FROM --platform=$BUILDPLATFORM node:18-alpine AS web_image

# 使用淘宝npm镜像源加速依赖安装
RUN npm config set registry https://registry.npmmirror.com
RUN npm install pnpm -g
# 配置 pnpm 使用淘宝镜像源
RUN pnpm config set registry https://registry.npmmirror.com

WORKDIR /build

# 先复制依赖文件（利用 Docker 缓存层）
COPY package.json package-lock.json pnpm-lock.yaml ./
RUN pnpm install

# 再复制其他文件并构建
COPY . .
RUN pnpm run build

# build backend
FROM --platform=$BUILDPLATFORM golang:1.21-bullseye AS server_image

# 【修复 1】必须显式声明 ARG TARGETARCH 才能获取 buildx 传入的架构参数
ARG TARGETARCH

WORKDIR /build
COPY ./service .

# 使用阿里云镜像源加速apt安装 (Debian Bullseye)
RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list \
    && sed -i 's/security.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list \
    || true

# 安装交叉编译工具链 (Debian/Ubuntu 基础镜像)
RUN apt-get update && apt-get install -y --no-install-recommends \
    bash curl gcc git musl-dev \
    && if [ "${TARGETARCH}" = "arm64" ]; then \
        apt-get install -y --no-install-recommends gcc-aarch64-linux-gnu libc6-dev-arm64-cross; \
    fi \
    && rm -rf /var/lib/apt/lists/*

# 中国国内源 (根据需要启用)
RUN go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct

RUN go install github.com/go-bindata/go-bindata/v3/go-bindata@latest

RUN rm -f bindata.go assets/bindata.go \
    && /go/bin/go-bindata -o=assets/bindata.go -pkg=assets -ignore="bindata.go" assets/... \
    && if [ "${TARGETARCH}" = "arm64" ]; then \
        CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=${TARGETARCH} go build -ldflags="-s -w -X sun-panel/global.RUNCODE=release -X sun-panel/global.ISDOCKER=docker" -o sun-panel main.go; \
    else \
        CGO_ENABLED=1 GOOS=linux GOARCH=${TARGETARCH} go build -ldflags="-s -w -X sun-panel/global.RUNCODE=release -X sun-panel/global.ISDOCKER=docker" -o sun-panel main.go; \
    fi


# run_image
# 【修复 2】将运行镜像改为 debian:bullseye-slim，与编译环境的 glibc 保持一致，避免 CGO 运行报错
FROM debian:bullseye-slim

WORKDIR /app

COPY --from=web_image /build/dist /app/web
COPY --from=server_image /build/sun-panel /app/sun-panel

EXPOSE 3002

# 更新使用 apt-get 安装运行时依赖
RUN apt-get update && apt-get install -y --no-install-recommends \
    bash ca-certificates tzdata \
    && chmod +x ./sun-panel \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && mkdir -p /data/conf /data/database /data/uploads /data/runtime \
    && rm -rf /var/lib/apt/lists/*

CMD ["./sun-panel"]
