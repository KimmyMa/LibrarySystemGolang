package models

type Lend struct {
	SerNum   int64     `gorm:"primaryKey;autoIncrement:true" json:"ser_num"`
	BookID   int64     `json:"book_id"`
	ReaderID int64     `json:"reader_id"`
	LendDate LocalDate `json:"lend_date"`
	BackDate LocalDate `json:"back_date"`

	// 定义外键关系
	Book       Book       `gorm:"foreignKey:BookID;references:BookID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ReaderInfo ReaderInfo `gorm:"foreignKey:ReaderID;references:ReaderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
