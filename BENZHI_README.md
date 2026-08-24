# EpochClock

EpochClock 是一个全局时序与单调序列服务。它为分布式节点分配单调递增的序列号，
通过区间批量分配与 WAL 持久化保证重启后序列不回退；节点以签发租约持有签发权，
租约过期或故障转移后签发权自动收敛；服务内置混合逻辑时钟（HLC），在物理时钟
回拨时仍保持时间戳单调。

## 构建

```bash
go build -mod=vendor ./...
```

依赖已 vendor 离线打包，构建时不需要访问外网。

## 运行

```bash
go run ./cmd/epochclock -addr :8080 -data ./data
```

启动后可用以下地址访问：

- `GET /healthz`：健康检查，返回 200。
- `GET /api/status`：节点、租约、序列区间、时钟与指标快照。
- `GET /monitor`：浏览器时序监控页面。
- `POST /v1/sequences/issue`：签发请求，JSON 请求体包含 `node_id`、`lease_id`、
  `count`，返回单调序号区间与时间戳。

## 目录

- `cmd/epochclock`：启动入口与 HTTP 服务。
- `internal/alloc`：序列区间分配。
- `internal/clock`：混合逻辑时钟。
- `internal/lease`：签发租约管理。
- `internal/persist`：WAL 持久化与重放。
- `internal/node`：节点注册。
- `internal/metric`：运行指标。
- `internal/model`：领域模型。
- `web`：浏览器监控页面。
