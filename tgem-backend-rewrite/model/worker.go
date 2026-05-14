package model

type Worker struct {
	ID                uint   `json:"id" gorm:"primaryKey"`
	ProjectID         uint   `json:"projectID"`
	Name              string `json:"name" gorm:"tinyText"`
	CompanyWorkerID   string `json:"companyWorkerID"`
	JobTitleInCompany string `json:"jobTitleInCompany"`
	JobTitleInProject string `json:"jobTitleInProject" gorm:"tinyText"`
	MobileNumber      string `json:"mobileNumber" gorm:"tinyText"`

	User User `json:"-" gorm:"foreignKey:WorkerID"`

	//Object Workers
	ObjectSupervisors []ObjectSupervisors `json:"-" gorm:"foreignKey:SupervisorWorkerID"`

	//Team Leaders
	TeamLeaderss []TeamLeaders `json:"-" gorm:"foreignKey:LeaderWorkerID"`

	//Invoice Input Workers
	InvoiceInputsWarehouseManager []InvoiceInput `json:"-" gorm:"foreignKey:WarehouseManagerWorkerID"`
	InvoiceInputsReleased         []InvoiceInput `json:"-" gorm:"foreignKey:ReleasedWorkerID"`

	//Invoice Object
	InvoiceObjectsSupervisor []InvoiceObject `json:"-" gorm:"foreignKey:SupervisorWorkerID"`

	//Invoice Return
	InvoiceReturns []InvoiceReturn `json:"-" gorm:"foreignKey:AcceptedByWorkerID"`

	//Invoice Output
	InvoiceOutputsWarehouseManager []InvoiceOutput `json:"-" gorm:"foreignKey:WarehouseManagerWorkerID"`
	InvoiceOutputsReleased         []InvoiceOutput `json:"-" gorm:"foreignKey:ReleasedWorkerID"`
	InvoiceOutputsRecipient        []InvoiceOutput `json:"-" gorm:"foreignKey:RecipientWorkerID"`

	//Invoice Output out of project
	InvoiceOutputOutOfProjectReleased []InvoiceOutputOutOfProject `json:"-" gorm:"foreignKey:ReleasedWorkerID"`

	//Invoice Object Operators
	InvoiceObjectOperators []InvoiceObjectOperator `json:"-" gorm:"foreignKey:OperatorWorkerID"`

	//Invoice WriteOff Relased
	InvoiceWriteOffReleaseds []InvoiceWriteOff `json:"-" gorm:"foreignKey:ReleasedWorkerID"`

	WorkerAttendances []WorkerAttendance `json:"-" gorm:"foreignKey:WorkerID"`
}
