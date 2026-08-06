# build frontend
FROM --platform=$BUILDPLATFORM docker.1ms.run/library/node:18-alpine AS web_image

RUN npm config set registry https://registry.npmmirror.com
RUN npm install pnpm -g
RUN pnpm config set registry https://registry.npmmirror.com

WORKDIR /build

COPY package.json package-lock.json pnpm-lock.yaml ./
RUN pnpm install

COPY . .
RUN pnpm run build


# build backend
# 最新alpine3.19导致sqlite3编译失败(https://github.com/mattn/go-sqlite3/issues/1164，
# 临时解决方案:https://github.com/mattn/go-sqlite3/pull/1177)
# sun-panel暂时解决方案使用golang:1.21-alpine3.18（因旧版本使用没问题，短期内较稳定） 
FROM docker.1ms.run/library/golang:1.21-alpine3.18 AS server_image

WORKDIR /build
COPY ./service .

# 【优化 2】在 Alpine 中安装原生的 CGO 编译依赖 (gcc, musl-dev)
RUN apk add --no-cache bash gcc musl-dev git

RUN go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct

RUN go install github.com/go-bindata/go-bindata/v3/go-bindata@latest

# 【优化 3】去掉复杂的 if/else 架构判断，因为当前容器已经是目标架构，直接 build 即可
RUN rm -f bindata.go assets/bindata.go \
    && /go/bin/go-bindata -o=assets/bindata.go -pkg=assets -ignore="bindata.go" assets/... \
    && CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w -X sun-panel/global.RUNCODE=release -X sun-panel/global.ISDOCKER=docker" -o sun-panel main.go


# run_image
# 【优化 4】恢复使用体积最小的 alpine 作为最终运行环境
FROM docker.1ms.run/library/alpine:latest

WORKDIR /app

COPY --from=web_image /build/dist /app/web
COPY --from=server_image /build/sun-panel /app/sun-panel

EXPOSE 3002

RUN apk add --no-cache bash ca-certificates su-exec tzdata \
    && chmod +x ./sun-panel \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && mkdir -p /data/conf /data/database /data/uploads /data/runtime

CMD ["./sun-panel"]
