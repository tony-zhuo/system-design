# Elevator System

## Problem Statement

設計一個電梯系統，能夠高效地處理多樓層建築中的乘客請求。從單部電梯的基本狀態機開始，逐步擴展到 SCAN 排程、多電梯調度，最後討論進階需求。

本題目採用**漸進式關卡**設計，每一關在前一關的基礎上擴展：

| Level | 主題 | 核心概念 |
|-------|------|----------|
| 1 | 單部電梯狀態機 | OOP 建模、State Pattern、基本請求處理 |
| 2 | SCAN 排程演算法 | 區分 hall call / cab call、LOOK algorithm |
| 3 | 多部電梯調度 | Dispatcher 策略模式、最佳電梯選擇 |
| 4 | 進階需求（設計題延伸） | 超重偵測、VIP 樓層、尖峰優化、維護模式 |

## Requirements

### Functional

**Level 1 — 單部電梯狀態機**
- 電梯可在指定樓層範圍內移動
- 接受乘客請求（目標樓層）
- 模擬逐層移動、開門、關門的過程
- 到達目標樓層時自動開門

**Level 2 — SCAN 排程**
- 區分 Hall Call（外部按鈕，帶方向）與 Cab Call（內部按鈕，只有目標樓層）
- 使用 LOOK 演算法：先處理同方向的請求，到底再反轉
- 不走到最高/最低樓層才反轉，而是到最遠的請求即反轉

**Level 3 — 多電梯調度**
- 管理 N 部電梯
- Hall Call 由 Dispatcher 分配給最佳電梯
- 考量距離、方向一致性、負載均衡

**Level 4 — 進階需求（Follow-up，未實作）**
- 超重偵測：電梯超重時不再接客
- VIP 樓層：特定樓層有優先權
- 尖峰時段優化：早上集中往上、傍晚集中往下
- 維護模式：電梯可進入維護狀態，不接受新請求

### Non-Functional
- 單次 `Step()` 呼叫 O(1) 時間複雜度
- 支援即時模擬（每個 step 代表一個時間單位）
- 設計易於擴展新的排程策略

## High-Level Design

```
┌─────────────────────────────────────────────┐
│                 Dispatcher                  │
│  - elevators []*Elevator                    │
│  - Dispatch(Request) → *Elevator            │
│  - StepAll() → advance all elevators        │
│  - cost(elevator, request) → float64        │
└──────────┬──────────┬──────────┬────────────┘
           │          │          │
     ┌─────▼──┐ ┌─────▼──┐ ┌────▼───┐
     │ Elev 1 │ │ Elev 2 │ │ Elev 3 │
     │        │ │        │ │        │
     │ state  │ │ state  │ │ state  │
     │ floor  │ │ floor  │ │ floor  │
     │ dir    │ │ dir    │ │ dir    │
     │ stops  │ │ stops  │ │ stops  │
     └────────┘ └────────┘ └────────┘
```

### 類別關係

```
Direction  (enum: Idle, Up, Down)
ElevatorState (enum: Idle, MovingUp, MovingDown, DoorOpen)
RequestType (enum: HallCall, CabCall)
Request { Floor, Direction, Type }

Elevator {
    ID, CurrentFloor, State, Direction
    upStops   []bool   // LOOK: 往上的停靠點
    downStops []bool   // LOOK: 往下的停靠點
    minRequest, maxRequest int  // 快取邊界，O(1) 判斷方向
    AddRequest(Request)
    Step() string
}

Dispatcher {
    Elevators []*Elevator
    Dispatch(Request) *Elevator
    StepAll() []string
}
```

## Detailed Design

### Level 1 — 電梯狀態機

狀態轉移圖：

```
        AddRequest
  Idle ──────────► MovingUp / MovingDown
   ▲                      │
   │                      │ arrived at target
   │                      ▼
   │              DoorOpen (timer=2)
   │                      │
   └──────────────────────┘
         timer expired, no more requests
```

- 每次 `Step()` 做一件事：移動一層 **或** 處理開關門
- 門開啟後維持 `doorOpenSteps`（2 步），然後關門
- 關門後由 `pickDirection()` 決定下一步

### Level 2 — LOOK Algorithm

LOOK 是 SCAN 的變體，差異在於不需要走到最頂/最底樓層才反轉：

```
假設電梯在 1F，方向 Up，請求: 3, 7, 5, 2(Down)

upStops:   {3, 5, 7}    ← cab calls 和 HallCall(Up)
downStops: {2}           ← HallCall(Down)

移動順序: 1→3→5→7 (到頂，反轉) →2
```

**請求分類邏輯：**
- `HallCall(Up)` → 放入 `upStops`
- `HallCall(Down)` → 放入 `downStops`
- `CabCall` → 根據相對位置放入對應的 stop set

**反轉條件：**
- 當前方向沒有更多請求時，切換到反方向
- 如果兩個方向都沒有請求，進入 Idle

### Level 3 — Dispatcher Cost Function

Dispatcher 使用 cost function 選擇最佳電梯：

```
cost(elevator, request):
    base = |currentFloor - requestFloor|

    if elevator is idle:
        return base + 0.5 * pendingCount

    if moving toward request AND same direction:
        return base + 0.5 * pendingCount        ← 最佳情況

    if moving toward BUT opposite direction:
        return base + span/2 + 0.5 * pendingCount

    if moving away:
        return detour_distance + 0.5 * pendingCount
```

**選擇優先順序：**
1. 同方向且順路的電梯（cost 最低）
2. 閒置的電梯（純距離）
3. 反方向或遠離的電梯（需要繞路）

`0.5 * pendingCount` 的負載權重確保請求不會集中在同一部電梯。

### Level 4 — 進階需求（Follow-up 題目）

以下為設計討論題，**尚未實作**，可作為練習延伸：

#### 4.1 超重偵測
- `Elevator` 新增 `currentWeight` 和 `maxWeight` 欄位
- `Step()` 中開門時檢查重量，超重則關門不載客
- 可用 `WeightSensor` interface 抽象感測器

#### 4.2 VIP 樓層
- `Request` 新增 `Priority` 欄位
- VIP 請求插入 stop set 時優先處理
- 可用 priority queue 取代 map

#### 4.3 尖峰時段優化
- 早上：多數請求從 1F 往上 → 電梯閒置時回到 1F 待命
- 傍晚：多數請求往 1F → 電梯閒置時分散到高樓層待命
- `Dispatcher` 根據時段調整 idle elevator 的預設位置

#### 4.4 維護模式
- `Elevator` 新增 `MaintenanceMode` flag
- 進入維護模式：完成當前請求後不再接受新請求
- `Dispatcher` 排除維護中的電梯

## Stop Set 資料結構比較

本專案實作了三種 stop set 資料結構，皆使用相同的 LOOK 排程邏輯，方便比較取捨。

### 檔案對應

| 檔案 | Struct | 資料結構 |
|------|--------|----------|
| `elevator.go` | `Elevator` | `[]bool` + `minRequest` / `maxRequest` 快取 |
| `elevator_bitmask.go` | `BitmaskElevator` | `uint64` 位元遮罩 |
| `elevator_bitset.go` | `BitsetElevator` | `github.com/bits-and-blooms/bitset` |

### 操作複雜度

| 操作 | `[]bool` + min/max | `uint64` bitmask | `bitset` 套件 |
|------|-------------------|-----------------|--------------|
| 查詢某層是否停靠 | O(1) `stops[i]` | O(1) `bits & (1<<i)` | O(1) `.Test(i)` |
| 設定 / 取消停靠 | O(1) `stops[i] = true/false` | O(1) `bits \|= / &^=` | O(1) `.Set()` / `.Clear()` |
| hasStopsAbove | O(1) `maxRequest > currentFloor` | O(1) bit mask | O(1) `.NextSet()` |
| hasStopsBelow | O(1) `minRequest < currentFloor` | O(1) bit mask | O(1) `.NextSet()` |
| HasPendingRequests | O(1) `minRequest <= maxRequest` | O(1) `bits != 0` | O(1) `.Any()` |
| PendingCount | O(n) 遍歷 array | O(1) `bits.OnesCount64` | O(n/64) `.Count()` |
| 移除邊界樓層時 | O(n) `recalcBounds` 重新掃描 | O(1) 無額外成本 | O(1) 無額外成本 |

> n = 樓層數

### 記憶體使用

| | `[]bool` | `uint64` bitmask | `bitset` 套件 |
|---|---------|-----------------|--------------|
| 每組 stops | n bytes | 8 bytes | n/8 bytes + struct overhead |
| 兩組合計（10 層） | 20 bytes | 16 bytes | ~48 bytes（含 struct） |
| 兩組合計（50 層） | 100 bytes | 16 bytes | ~32 bytes |
| 兩組合計（100 層） | 200 bytes | 不支援（上限 64 層） | ~48 bytes |

### 優缺點分析

#### `[]bool` + minRequest / maxRequest（`elevator.go`）

**適合場景：** 面試白板題、教學、中小型建築

```go
// 優勢：最直覺，O(1) 方向判斷
func (e *Elevator) hasStopsAbove() bool {
    return e.maxRequest > e.CurrentFloor
}

// 代價：停靠在邊界樓層時需要 O(n) recalcBounds
func (e *Elevator) openDoor() {
    // ...
    if e.CurrentFloor == e.minRequest || e.CurrentFloor == e.maxRequest {
        e.recalcBounds() // O(n) 但不頻繁
    }
}
```

- 無外部依賴、無樓層數限制
- `PendingCount` 仍為 O(n)
- `recalcBounds` 只在停靠邊界樓層時觸發，攤提成本低

#### `uint64` bitmask（`elevator_bitmask.go`）

**適合場景：** 追求極致效能、嵌入式系統、64 層以下建築

```go
// 所有操作都是 O(1)，無任何 rescan
func (e *BitmaskElevator) hasStopsAbove() bool {
    mask := aboveMask(e.idx(e.CurrentFloor)) // ^uint64(0) << (bit+1)
    return (e.upStops|e.downStops)&mask != 0
}

func (e *BitmaskElevator) PendingCount() int {
    return bits.OnesCount64(e.upStops) + bits.OnesCount64(e.downStops)
}
```

- 所有操作真正 O(1)，包括 `PendingCount`（CPU popcount 指令）
- 記憶體最小：兩個 `uint64` = 16 bytes
- Cache 友善：整個 stop set 在一個 CPU word
- **限制：最多 64 層**，超過需改用 `[]uint64` 多 word 方案

#### `bitset` 套件（`elevator_bitset.go`）

**適合場景：** 超高層建築（100+ 層）、需要豐富 API

```go
// NextSet 找到下一個 set bit，不需要逐層掃描
func (e *BitsetElevator) hasStopsAbove() bool {
    start := e.idx(e.CurrentFloor) + 1
    if next, ok := e.upStops.NextSet(start); ok && next < n { return true }
    if next, ok := e.downStops.NextSet(start); ok && next < n { return true }
    return false
}
```

- 無樓層數限制，自動管理底層 `[]uint64`
- 提供 `NextSet` / `Count` / `Union` / `Intersection` 等豐富操作
- **需要外部依賴** `github.com/bits-and-blooms/bitset`
- 比手寫 bitmask 多一層抽象開銷

### 選擇建議

```
樓層數 ≤ 64 且追求效能 → uint64 bitmask
樓層數 > 64 或需要集合運算 → bitset 套件
面試白板 / 教學 / 快速原型  → []bool + min/max
```

## Trade-offs & Alternatives

| 決策 | 選擇 | 替代方案 | 理由 |
|------|------|----------|------|
| 排程演算法 | LOOK | FCFS / Shortest Seek First | LOOK 兼顧公平性與效率，避免 starvation |
| Stop set 資料結構 | `[]bool` + min/max 快取 | `uint64` bitmask / `bitset` 套件 | 三種皆實作，詳見上方比較 |
| 調度策略 | Cost function | Round Robin / Zone-based | Cost function 可彈性調整權重，適合面試討論 |
| 時間模擬 | 離散 Step | 事件驅動 (event queue) | Step-based 更直覺，易於測試和 debug |
| 門開啟時間 | 固定 2 步 | 可配置 / 動態調整 | 簡化設計，Level 4 可擴展 |

## References

- [LOOK Disk Scheduling Algorithm](https://en.wikipedia.org/wiki/LOOK_algorithm) — 電梯排程的原型
- [Elevator Algorithm](https://en.wikipedia.org/wiki/Elevator_algorithm) — SCAN / C-SCAN 變體
- [System Design: Elevator System](https://www.geeksforgeeks.org/system-design-elevator-system/) — 系統設計參考
