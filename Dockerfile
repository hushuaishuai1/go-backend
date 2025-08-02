# --- Build Stage ---
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 预先复制依赖文件，利用Docker缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源代码
COPY . .

# 编译应用
# CGO_ENABLED=0 是为了静态编译，GOOS=linux是为了确保在Linux上运行
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-app ./cmd

# --- Final Stage ---
FROM alpine:3.18

# SQLite需要libc
RUN apk add --no-cache libc6-compat

WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /go-app /app/go-app

# 数据库文件将通过volume挂载到/app/data
VOLUME /app/data

EXPOSE 8080

CMD ["./go-app"]