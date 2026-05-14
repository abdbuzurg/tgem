package model

type Object struct {
	ID uint `json:"id" gorm:"primaryKey"`

	//Custom field that references different type of Object tables in the database
	//The type will be determined by type filed on the application level not database
	ObjectDetailedID uint   `json:"objectDetailedID"`
	Type             string `json:"type" gorm:"tinyText"`
	Name             string `json:"name" gorm:"tinyText"`
	Status           string `json:"status" gorm:"tinyText"`
	ProjectID        uint   `json:"projectID"`

	ObjectTeams       []ObjectTeams       `json:"-" gorm:"foreignKey:ObjectID"`
	ObjectSupervisors []ObjectSupervisors `json:"-" gorm:"foreignKey:ObjectID"`

	InvoiceObject []InvoiceObject `json:"-" gorm:"foreignKey:ObjectID"`

	TPNourashesObjects []TPNourashesObjects `json:"-" gorm:"foreignKey:TP_ObjectID"`

  SubstationObjects     []SubstationCellNourashesSubstationObject `json:"-" gorm:"foreignKey:SubstationObjectID"`
	SubstationCellObjects []SubstationCellNourashesSubstationObject `json:"-" gorm:"foreignKey:SubstationCellObjectID"`
}

type MJD_Object struct {
	ID              uint   `json:"id" gorm:"primaryKey"`
	Model           string `json:"model" gorm:"tinyText"`
	AmountStores    uint   `json:"amountStores"`
	AmountEntrances uint   `json:"amountEntrances"`
	HasBasement     bool   `json:"hasBasement"`
}

type TP_Object struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	Model        string `json:"model" gorm:"tinyText"`
	VoltageClass string `json:"voltageClass" gorm:"tinyText"`
}

type TPNourashesObjects struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	TP_ObjectID uint   `json:"tpObjectID"`
	TargetID    uint   `json:"targetID"`
	TargetType  string `json:"targetType"`
}

type STVT_Object struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	VoltageClass  string `json:"voltageClass" gorm:"tinyText"`
	TTCoefficient string `json:"ttCoefficient" gorm:"tinyText"`
}

type SIP_Object struct {
	ID            uint `json:"id" gorm:"primaryKey"`
	AmountFeeders uint `json:"amountFeeders"`
}

type KL04KV_Object struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	Length    float64 `json:"length"`
	Nourashes string  `json:"nourashes" gorm:"tinyText"`
}

type Substation_Object struct {
	ID                   uint   `json:"id" gorm:"primaryKey"`
	VoltageClass         string `json:"voltageClass"`
	NumberOfTransformers uint   `json:"numberOfTransformers"`
}

type SubstationCellObject struct {
	ID uint `gorm:"primaryKey"`
}

type SubstationCellNourashesSubstationObject struct {
	ID                     uint `gorm:"primaryKey"`
	SubstationObjectID     uint
	SubstationCellObjectID uint
}
