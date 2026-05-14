export interface IInvoiceObject {
	id: number
	deliveryCode: string 
  districtid: number
	projectID: number
	supervisorWorkerID: number
  dateOfInvoice: Date
	objectID: number
	teamID: number
}

export interface IInvoiceObjectMaterials {
  materialID: number
  materialName: string
  availableMaterial: number
  unit: string
  amount: number
  notes: string
  hasSerialNumbers: boolean
  serialNumbers: string[]
}

export interface IInvoiceObjectOperations {
  operationID: number
  operationName: string
  amount: number
  notes: string
}
