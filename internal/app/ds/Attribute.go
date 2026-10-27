package ds

type Attribute struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"type:varchar(100);not null" json:"name"`
}

type AttributeValue struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	SwadeshListID  uint      `gorm:"not null" json:"swadesh_list_id"`
	AttributeID    uint      `gorm:"not null" json:"attribute_id"`
	Value          string    `gorm:"type:varchar(255)" json:"value"`

	Attribute      Attribute `gorm:"foreignKey:AttributeID" json:"attribute"`
}
