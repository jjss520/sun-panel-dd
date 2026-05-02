# build frontend
FROM docker.m.daocloud.io/node:18-alpine AS web_image

# 使用淘宝npm镜像源加速依赖安装
RUN npm config set registry https://registry.npmmirror.com

RUN npm install pnpm -g

# 配置 pnpm 使用淘宝镜像源
RUN pnpm config set registry https://registry.npmmirror.com

WORKDIR /build

# 先复制依赖文件（利用 Docker 缓存层）
COPY package.json package-lock.json pnpm-lock.yaml ./

# 安装依赖
RUN pnpm install

# 再复制其他文件
COPY . .

# 构建项目
RUN pnpm run build

# build backend
# sun-panel暂时解决方案使用golang:1.21-alpine3.18（因旧版本使用没问题，短期内较稳定）
FROM docker.m.daocloud.io/golang:1.21-alpine3.18 AS server_image

WORKDIR /build

COPY ./service .

# 使用阿里云镜像源加速apk安装
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk add --no-cache bash curl gcc git musl-dev

# 中国国内源 (根据需要启用)
RUN go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct

RUN go install github.com/go-bindata/go-bindata/v3/go-bindata@latest

RUN rm -f bindata.go assets/bindata.go \
    && /go/bin/go-bindata -o=assets/bindata.go -pkg=assets -ignore="bindata.go" assets/... \
    && go build -o sun-panel --ldflags="-X sun-panel/global.RUNCODE=release -X sun-panel/global.ISDOCKER=docker" main.go



# run_image
FROM docker.m.daocloud.io/alpine:latest

WORKDIR /app

COPY --from=web_image /build/dist /app/web

COPY --from=server_image /build/sun-panel /app/sun-panel

# 中国国内源
# RUN sed -i "s@dl-cdn.alpinelinux.org@mirrors.aliyun.com@g" /etc/apk/repositories

EXPOSE 3002

RUN apk add --no-cache bash ca-certificates su-exec tzdata \
    && chmod +x ./sun-panel \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && mkdir -p /data/conf /data/database /data/uploads /data/runtime

CMD ["./sun-panel"]
