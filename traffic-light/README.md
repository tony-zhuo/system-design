# Traffic Signal System

📋 題目
> 設計一個城市交通號誌控制系統（Traffic Light Control System）
> 系統需要管理一座城市中的所有紅綠燈，支援智慧調度（根據車流量動態調整），並在號誌故障時能自動偵測與告警。

## Functional Requirements
- 管理 10,000 個路口、40,000 個號誌燈的狀態
- 固定時序模式 + 動態車流調整模式
- 緊急車輛優先通行
- 中控室人工覆寫
- 故障偵測與告警

## Non-Functional Requirements
- Fail-Safe：故障預設 All Red
- 高可用性（號誌是安全關鍵）
- 保留 90 天歷史資料

```text
[路口層]
  Traffic Signals × N
       ↑ 控制指令
  Controller (Primary)  ←→  Controller (Standby)
       ↓ 狀態回報 / 事件

[接入層]
  API Gateway (Protocol Bridge)
       ↓ 標準化後
  Message Queue (Kafka)
       ↓
  Consumer Service → DB (歷史資料)

[控制層]
  Scheduling Service（跨區綠波協調）
       ↓ 調度指令
  API Gateway → Controller
  
[介面層]
  中控室 Dashboard
  緊急車輛系統
```

層級          故障情境              處理方式
─────────────────────────────────────────────
硬體層        號誌燈壞掉             主幹道閃黃、支線閃紅
             Controller 掛掉       Standby 接管（等待確認避免 Split-Brain）
網路層        有線斷線               自動切換 4G/5G LTE
服務層        Backend Server 掛     Load Balancer 切換
             Kafka 掛              Replication 保障不遺失資料
             DB 掛                 Auto Failover（Patroni）