package models

type Book struct {
	BookID       int64   `gorm:"primaryKey;autoIncrement:true" json:"book_id"`
	Name         string  `json:"name"`
	Author       string  `json:"author"`
	Publish      string  `json:"publish"`
	ISBN         string  `json:"isbn"`
	Introduction string  `json:"introduction"`
	Language     string  `json:"language"`
	Price        float64 `json:"price"`
	PubDate      string  `json:"pub_date"`
	ClassID      int64   `json:"class_id"`
	Number       int     `json:"number"`
	Image        string  `json:"image"`

	// 定义外键关系
	ClassInfo ClassInfo `gorm:"foreignKey:ClassID;references:ClassID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
