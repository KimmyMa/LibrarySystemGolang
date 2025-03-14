package models

type Reserve struct {
	SerNum      int64     `gorm:"primaryKey;autoIncrement:true" json:"ser_num"`
	BookID      int64     `json:"book_id"`
	ReaderID    int64     `json:"reader_id"`
	RequireDate LocalDate `json:"require_date"`
	AcceptDate  LocalDate `json:"accept_date"`

	// 定义外键关系
	Book       Book       `gorm:"foreignKey:BookID;references:BookID"`
	ReaderInfo ReaderInfo `gorm:"foreignKey:ReaderID;references:ReaderID"`
}
