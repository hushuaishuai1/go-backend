package models

import (
	"gorm.io/gorm"
)

// Alert 定义了警报的数据结构
type Alert struct {
	gorm.Model // 包含了ID, CreatedAt, UpdatedAt, DeletedAt
	Email       string  `gorm:"index"` // 为邮箱创建索引以加快查询
	Symbol      string  `gorm:"index"` // 为交易对创建索引
	TargetPrice float64
	Direction   string // "ABOVE" or "BELOW"
}

// CreateAlertRequest 定义了创建警报API的请求体
type CreateAlertRequest struct {
	Email       string  `json:"email" binding:"required,email"`
	Symbol      string  `json:"symbol" binding:"required"`
	TargetPrice float64 `json:"targetPrice" binding:"required,gt=0"`
	Direction   string  `json:"direction" binding:"required,oneof=ABOVE BELOW"`
}