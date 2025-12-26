# Zeus API 文档

- 基础地址: `http://<host>:<port>`（端口来自配置 `http.port`）
- 认证: 如果配置项 `http.auth` 不为空，所有非 `/_internal/*` 路由都必须在请求头携带 `Authorization: Bearer <token>`
- 所有请求/响应均使用 `application/json`
- 错误响应统一为 `{"error": string}`，除非特别说明（Echo 的默认错误结构除外）

## 健康检查

- 方法/路径: `GET /_internal/health`
- 成功响应: `200` `{ "message": "ok" }`
- 失败响应: `500` `{ "error": string }`

## 连接检查

- 方法/路径: `GET /connection/check`
- 认证: 需要（与全局 Bearer 认证一致）
- 成功响应: `200` `{ "ok": true }`
- 失败响应: `401` `{ "message": string }`（当认证失败时，由 Echo 返回）

## 资源: 池 (Pool)

### 获取池使用信息
- 方法/路径: `GET /pool/info`
- 响应: `200` `PoolInfo`（按 region 聚合的容量与使用量）

TypeScript: `type PoolInfo = Record<string, { size: number; used: number; friendlyName: string }>`

### 创建地址池
- 方法/路径: `POST /pools`
- 请求体: `CreatePoolPayload`
  - `start: number` 起始地址（整型表示）
  - `gateway: number` 网关地址（整型表示）
  - `size: number` 池大小
  - `region: string` 区域标识
- 响应: `200` `{ id: string }`
  - 失败: `400` `{ "error": string }`（当 `region` 不存在时）

### 删除地址池
- 方法/路径: `DELETE /pool/:id`
- 路径参数: `id: string`
- 响应:
  - 成功: `204 No Content`
  - 已分配地址存在或仍有租约: `409` `{ "error": string }`
  - 池不存在: `404` `{ "error": string }`
  - 缺少 `id`: `400` `{ "error": string }`

### 按区域获取池详情
- 方法/路径: `GET /pool/region/:region`
- 路径参数: `region: string`
- 响应: `200` `PoolDetail[]`
  - `id: string`
  - `region: string`
  - `friendlyName: string`
  - `begin: string` 起始 IP（点分十进制）
  - `end: string` 结束 IP（点分十进制）
  - `gateway: string` 网关 IP
  - `state: AllocateState[]` 分配状态数组（与池大小一致，取值 0/1/2）

### 按池 ID 获取详情
- 方法/路径: `GET /pool/id/:id`
- 路径参数: `id: string`
- 响应: `200` `PoolDetail`
  - 字段同上（`PoolDetail`）

## 资源: 区域 (Region)

### 列出区域
- 方法/路径: `GET /regions`
- 响应: `200` `Region[]`
  - 返回字段：`id`, `name`, `friendlyName`, `createdAt`

### 创建区域
- 方法/路径: `POST /regions`
- 请求体: `CreateRegionPayload`
  - `name: string`
  - `friendlyName: string`
- 响应: `200` `{ id: string }`
  - 名称冲突: `409` `{ "error": string }`

### 更新区域
- 方法/路径: `PATCH /region/:id`
- 路径参数: `id: string`
- 请求体: `UpdateRegionPayload`
  - `name?: string`
  - `friendlyName?: string`
- 响应: `200` `Region`
  - 返回同上（字段为 `id`, `name`, `friendlyName`, `createdAt`）
  - 名称冲突: `409` `{ "error": string }`

### 删除区域
- 方法/路径: `DELETE /region/:id`
- 路径参数: `id: string`
- 响应:
  - 成功: `204 No Content`
  - 区域下仍存在池: `409` `{ "error": string }`
  - 区域不存在: `404` `{ "error": string }`
  - 缺少 `id`: `400` `{ "error": string }`

## 资源: 虚拟机 (VM)

### 创建虚拟机
- 方法/路径: `POST /vms`
- 请求体: `CreateVirtualMachineRequest`
  - `region: string[]`
  - `host: string`
  - `name: string`
  - `vmid: number`
  - `type: string`
- 响应: `200` `CreateVirtualMachineResponse`
  - `id: string`
  - `addresses: Record<string, AddressResult>`（按 region 返回分配的地址）

### 创建“迁入”占位（用于迁移）
- 方法/路径: `POST /vms/migration`
- 请求体: `MigrationVirtualMachineRequest`
  - `host: string`
  - `vmid: number`
  - `pools: string[]`
  - `index: number`
  - `type: string`
- 响应: `200` `{ id: string }`

### 查询指定主机上某个 VM
- 方法/路径: `GET /vm/:host/:id`
- 路径参数:
  - `host: string`
  - `id: number`（对应该主机上的 VM 数值 ID）
- 响应: `200` `VmInfo`

### 查询指定主机上的所有 VM
- 方法/路径: `GET /vm/:host`
- 路径参数: `host: string`
- 响应: `200` `VmInfo[]`

### 删除指定主机上的某个 VM
- 方法/路径: `DELETE /vm/:host/:id`
- 路径参数:
  - `host: string`
  - `id: number`（对应该主机上的 VM 数值 ID）
- 响应: `204 No Content`

说明：删除操作会通过 Assign 释放其所有租约并移除对应记录。

### 将 VM 标记为“迁出”（迁移开始）
- 方法/路径: `DELETE /vm/:host/:id/migration`
- 路径参数:
  - `host: string`
  - `id: number`
- 响应: `200` `VmInfo`（迁出后对象信息）

### 迁入到当前实例（完成迁移）
- 方法/路径: `PATCH /vm/:id`
- 路径参数: `id: string`（内部分配的 VM 字符串 ID）
- 请求体: `MigrateInRequest`
  - `host: string`
  - `vmid: number`
- 响应: `200` `VmInfo`

## 资源: 镜像 (Images)

注意：镜像接口为上游转发，响应字段由上游服务决定。

### 查询最新镜像
- 方法/路径: `GET /images/latest`
- 查询参数: `flavor: string`
- 响应: 透传上游 JSON（示例结构见 `SystemImage`）

### 查询指定镜像
- 方法/路径: `GET /image/:flavor/:version`
- 路径参数: `flavor: string`, `version: string`
- 响应: 透传上游 JSON（示例结构见 `SystemImage`）

## 资源: 端口 (Port)

### 分配端口
- 方法/路径: `POST /port`
- 请求体: `AssignPortRequest`
  - `host: string`（所属主机标识）
  - `virtualMachine: number`（主机上的 VM 数值 ID）
  - `targetPort: number`
  - `service: string`
- 响应: `200` `{ port: number }`

### 查询 VM 已分配端口
- 方法/路径: `GET /ports/:host/:id`
- 路径参数:
  - `host: string`
  - `id: number`（主机上的 VM 数值 ID）
- 响应: `200` `AssignResponse[]`

### 查询所有端口转发规则
- 方法/路径: `GET /ports`
- 响应: `200` `ForwardRule[]`
  - `inPort: number`
  - `outPort: number`
  - `outHost: string`
  - `proto: "tcp" | "udp"`

## 资源: 分配 (Assign)

### 按区域创建分配（批量分配地址）
- 方法/路径: `POST /assigns`
- 请求体: `AssignCreateRequest`
  - `region: string[]`
  - `host: string`
  - `key: string`（业务关键字，用于幂等/查找）
  - `type: string`
  - `data: object`（透传业务数据）
- 响应: `200` `AssignCreateResponse`
  - `id: string`
  - `addresses: Record<string, AddressResult>`（按 region 映射）

### 按池索引创建分配（固定 IP 索引）
- 方法/路径: `POST /assigns/index`
- 请求体: `AssignCreateIndexRequest`
  - `pools: string[]`
  - `index: number`
  - `host: string`
  - `key: string`
  - `type: string`
  - `data: object`
- 响应: `200` `{ id: string }`

### 按池索引查询分配
- 方法/路径: `GET /assigns/index`
- 查询参数: `pool: string`, `index: number`
- 响应: `200` `AssignInfo`
  - 未找到：`404` `{ "error": string }`

### 为已有分配添加区域地址
- 方法/路径: `POST /assign/:id/region`
- 路径参数: `id: string`
- 请求体: `{ region: string; host?: string }`
  - `region` 必填，指定需要追加的区域 ID
  - `host` 可选，未提供时会复用该 assign 已有地址的主机标识
- 响应: `200` `AssignInfo`
  - `409` `{ "error": "region already assigned" }` 当该区域已存在时
  - `400` `{ "error": "host is required" }` 当 assign 内没有可复用主机且未传 `host`
  - `404` `{ "error": "assign not found" }`

### 查询分配详情
- 方法/路径: `GET /assign/:id`
- 路径参数: `id: string`
- 响应: `200` `AssignInfo`
  - `id, createdAt, key, type, data, leases: Record<string, AddressResult>`

### 释放分配中的区域地址
- 方法/路径: `DELETE /assign/:id/region/:region`
- 路径参数: `id: string`, `region: string`
- 响应: `200` `AssignInfo`
  - 响应结果中移除指定区域的地址
  - `404` `{ "error": "assign region not found" }` 当 assign 不存在或不含该区域地址

### 删除分配
- 方法/路径: `DELETE /assign/:id`
- 路径参数: `id: string`
- 响应: `204 No Content`

---

## 主要数据结构（摘要）

以下是与 API 对应的核心结构（完整 TS 定义见 `types/api.d.ts`）。

- `AddressResult`: `{ address: string; gateway: string; leaseId: string; vlan?: number | null }`
- `BatchAddressResult`: `Record<string, AddressResult>`
- `VmInfo`:
  - `id: string`
  - `createdAt: string`（ISO 时间）
  - `host: string`
  - `leases: BatchAddressResult`
  - `name: string`
  - `vmid: number`
  - `type: string`
- `PoolInfo`: `Record<string, { size: number; used: number }>`
- `PoolDetail`:
  - `id: string`
  - `region: string`
  - `friendlyName: string`
  - `begin: string`
  - `end: string`
  - `gateway: string`
  - `state: AllocateState[]`
- `AllocateState`: `0 | 1 | 2`（0=未分配，1=已分配，2=不可用）
- `SystemImage`（上游示例）:
  - `id: string`
  - `createdAt: string`
  - `flavor: string`
  - `version: string`
  - `imageUrl: string`
