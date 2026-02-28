# System Design

練習各種 system design 題目，並以 Go 實作核心邏輯。

## 題目索引

| Topic | Status | Description |
|-------|--------|-------------|
| [elevator-system](elevator-system/) | Done | 電梯系統設計（單部狀態機 → SCAN 排程 → 多電梯調度） |

## 目錄結構

```
system-design/
├── CLAUDE.md
├── README.md
├── go.mod
├── pkg/                       # 跨題目共用的工具
└── <topic-name>/              # 各題目獨立目錄
```

### 題目目錄範本

每個題目依照以下結構組織：

```
<topic-name>/
├── README.md              # Design doc
│                          #   - Problem statement
│                          #   - Requirements (functional / non-functional)
│                          #   - High-level design
│                          #   - Detailed design
│                          #   - Trade-offs & alternatives
│                          #   - References
├── main.go                # 程式入口 / 可執行的 demo
├── <module>.go            # 核心實作（依需求拆分多個檔案）
└── <module>_test.go       # 對應的測試檔
```

### `pkg/` 共用套件

放置跨題目可復用的工具，例如：

- `pkg/hash/` — consistent hashing
- `pkg/logger/` — 統一 logging
- `pkg/testutil/` — 測試輔助函式
