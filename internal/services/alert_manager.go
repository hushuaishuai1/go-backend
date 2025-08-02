package services

import (
	"log"
	"sync"
	"github.com/hushuaishuai1/go-backend/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// AlertManager 负责管理所有警报
type AlertManager struct {
	db *gorm.DB
	// 使用 sync.Map 以保证并发安全
	// 结构: map[string]*sync.Map -> map[Symbol]*sync.Map
	// 第二层map: map[uint]models.Alert -> map[AlertID]Alert
	activeAlerts *sync.Map
}

func NewAlertManager(dbPath string) (*AlertManager, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// 自动迁移数据库结构
	db.AutoMigrate(&models.Alert{})

	manager := &AlertManager{
		db:           db,
		activeAlerts: &sync.Map{},
	}
	manager.loadAlertsFromDB()
	return manager, nil
}

// 从数据库加载所有警报到内存
func (am *AlertManager) loadAlertsFromDB() {
	var alerts []models.Alert
	am.db.Find(&alerts)
	for _, alert := range alerts {
		am.AddAlertToMemory(alert)
	}
	log.Printf("Loaded %d alerts from database into memory.", len(alerts))
}

// AddAlertToMemory 将单个警报添加到内存查找表
func (am *AlertManager) AddAlertToMemory(alert models.Alert) {
	symbolMap, _ := am.activeAlerts.LoadOrStore(alert.Symbol, &sync.Map{})
	symbolMap.(*sync.Map).Store(alert.ID, alert)
}

// RemoveAlertFromMemory 从内存中移除单个警报
func (am *AlertManager) RemoveAlertFromMemory(alert models.Alert) {
	if symbolMap, ok := am.activeAlerts.Load(alert.Symbol); ok {
		symbolMap.(*sync.Map).Delete(alert.ID)
	}
}

// CreateAlert 创建一个新警报并存入数据库和内存
func (am *AlertManager) CreateAlert(req models.CreateAlertRequest) (models.Alert, error) {
	alert := models.Alert{
		Email:       req.Email,
		Symbol:      req.Symbol,
		TargetPrice: req.TargetPrice,
		Direction:   req.Direction,
	}
	result := am.db.Create(&alert)
	if result.Error != nil {
		return models.Alert{}, result.Error
	}
	am.AddAlertToMemory(alert)
	log.Printf("Created and loaded new alert for %s on %s", alert.Email, alert.Symbol)
	return alert, nil
}

// GetAlertsByEmail 根据邮箱获取警报列表
func (am *AlertManager) GetAlertsByEmail(email string) ([]models.Alert, error) {
	var alerts []models.Alert
	result := am.db.Where("email = ?", email).Find(&alerts)
	return alerts, result.Error
}

// DeleteAlert 删除一个警报
func (am *AlertManager) DeleteAlert(alertID uint) error {
	var alert models.Alert
	// 先从数据库找到它，以便知道它的symbol
	if err := am.db.First(&alert, alertID).Error; err != nil {
		return err
	}
	
	// 从数据库删除
	result := am.db.Delete(&models.Alert{}, alertID)
	if result.Error == nil && result.RowsAffected > 0 {
		// 从内存中也删除
		am.RemoveAlertFromMemory(alert)
		log.Printf("Deleted alert ID %d for %s", alertID, alert.Symbol)
	}
	return result.Error
}

// GetAlertsBySymbol 从内存中高效获取某个交易对的所有警报
func (am *AlertManager) GetAlertsBySymbol(symbol string) []models.Alert {
	alerts := []models.Alert{}
	if symbolMap, ok := am.activeAlerts.Load(symbol); ok {
		symbolMap.(*sync.Map).Range(func(key, value interface{}) bool {
			alerts = append(alerts, value.(models.Alert))
			return true
		})
	}
	return alerts
}