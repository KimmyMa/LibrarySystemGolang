package models

type ClassInfo struct {
	ClassID   int64  `gorm:"primaryKey;autoIncrement:true" json:"class_id"`
	ClassName string `json:"class_name"`
}
