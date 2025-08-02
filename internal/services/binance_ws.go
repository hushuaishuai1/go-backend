package services

import (
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"
	"strconv"

	"github.com/gorilla/websocket"
	    "github.com/hushuaishuai1/go-backend/internal/models" // <-- 为解决问题2.2新增

)

const binanceWsURL = "wss://stream.binance.com:9443/ws"

// MiniTickerMessage 定义了我们从WebSocket接收的价格更新消息结构
type MiniTickerMessage struct {
	Event  string `json:"e"` // Event type
	Symbol string `json:"s"` // Symbol
	Close  string `json:"c"` // Close price
}

var subscribedSymbols sync.Map

// StartBinanceWebSocket 启动并维护与币安的WebSocket连接
func StartBinanceWebSocket(am *AlertManager, mailer *Mailer) {
	var conn *websocket.Conn
	var err error

	connect := func() {
		conn, _, err = websocket.DefaultDialer.Dial(binanceWsURL, nil)
		if err != nil {
			log.Printf("ERROR: Failed to connect to Binance WebSocket: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
		}
	}

	connect() // 首次连接

	go func() {
		for {
			if conn == nil {
				connect()
				continue
			}

			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("ERROR: WebSocket read error: %v. Reconnecting...", err)
				conn.Close()
				connect()
				continue
			}

			var msg MiniTickerMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				// 忽略非价格消息
				continue
			}
			
			// 高效地处理价格更新
			go handlePriceUpdate(msg, am, mailer)
		}
	}()

	// 定期同步需要订阅的交易对
	go func() {
		ticker := time.NewTicker(30 * time.Second) // 每30秒同步一次
		defer ticker.Stop()
		for range ticker.C {
			if conn == nil {
				continue
			}
			var allSymbols []string
			am.activeAlerts.Range(func(key, value interface{}) bool {
				allSymbols = append(allSymbols, key.(string))
				return true
			})
			updateSubscriptions(conn, allSymbols)
		}
	}()
}

// updateSubscriptions 更新WebSocket的订阅列表
func updateSubscriptions(conn *websocket.Conn, symbols []string) {
    newSubMap := make(map[string]bool)
    for _, s := range symbols {
        newSubMap[strings.ToLower(s)] = true
    }

    // Unsubscribe from symbols that are no longer needed
    var toUnsubscribe []string
    subscribedSymbols.Range(func(key, value interface{}) bool {
        if !newSubMap[key.(string)] {
            toUnsubscribe = append(toUnsubscribe, key.(string)+"@miniTicker")
            subscribedSymbols.Delete(key)
        }
        return true
    })
    if len(toUnsubscribe) > 0 {
        unsubscribeMsg := map[string]interface{}{
            "method": "UNSUBSCRIBE",
            "params": toUnsubscribe,
            "id":     1,
        }
        conn.WriteJSON(unsubscribeMsg)
    }

    // Subscribe to new symbols
    var toSubscribe []string
    for _, s := range symbols {
        symbolLower := strings.ToLower(s)
        if _, ok := subscribedSymbols.Load(symbolLower); !ok {
            toSubscribe = append(toSubscribe, symbolLower+"@miniTicker")
            subscribedSymbols.Store(symbolLower, true)
        }
    }
    if len(toSubscribe) > 0 {
        subscribeMsg := map[string]interface{}{
            "method": "SUBSCRIBE",
            "params": toSubscribe,
            "id":     1,
        }
        conn.WriteJSON(subscribeMsg)
    }
}


// handlePriceUpdate 当收到价格更新时，检查是否有警报需要触发
func handlePriceUpdate(msg MiniTickerMessage, am *AlertManager, mailer *Mailer) {
	price, err := strconv.ParseFloat(msg.Close, 64)
	if err != nil {
		return
	}
	
	alerts := am.GetAlertsBySymbol(msg.Symbol)
	for _, alert := range alerts {
		shouldTrigger := (alert.Direction == "ABOVE" && price > alert.TargetPrice) ||
			(alert.Direction == "BELOW" && price < alert.TargetPrice)

		if shouldTrigger {
			// 触发警报！
			log.Printf("TRIGGERED: %s price is %.4f, triggering alert for %s", alert.Symbol, price, alert.Email)
			// 异步发送邮件和删除，不阻塞价格处理
			go func(a models.Alert, p float64) {
				mailer.SendAlertEmail(a, p)
				am.DeleteAlert(a.ID)
			}(alert, price)
		}
	}
}