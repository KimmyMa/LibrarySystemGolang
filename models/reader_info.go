package models

type ReaderInfo struct {
	ReaderID int64     `gorm:"primaryKey;autoIncrement:true" json:"reader_id"`
	Name     string    `json:"name"`
	Sex      string    `json:"sex"`
	Birth    LocalDate `json:"birth"`
	Address  string    `json:"address"`
	Phone    string    `json:"phone"`
}
