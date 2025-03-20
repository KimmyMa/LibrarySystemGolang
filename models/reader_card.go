package models

type ReaderCard struct {
	ReaderID int64  `gorm:"primaryKey;autoIncrement:true" json:"reader_id"`
	Username string `json:"username"`
	Password string `json:"password"`

	// 定义外键关系
	ReaderInfo ReaderInfo `gorm:"foreignKey:ReaderID;references:ReaderID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}
