package model

type TestMigration struct {
  ID uint `gorm:"primaryKey"`
  Value int 
}
