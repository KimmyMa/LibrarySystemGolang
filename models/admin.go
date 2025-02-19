package models

type Admin struct {
	AdminID  int64  `gorm:"primaryKey;autoIncrement:true" json:"admin_id"`
	Password string `json:"password"`
	Username string `json:"username"`
}
