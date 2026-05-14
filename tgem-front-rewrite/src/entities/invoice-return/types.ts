export interface IInvoiceReturn {
  id: number
  projectID: number
  returnerID: number
  districtID: number
  returnerType: "object" | "team"
  acceptorID: number
  acceptorType: "team" | "warehouse"
  acceptedByWorkerID: number
  deliveryCode: string
  notes: string
  dateOfInvoice: Date
  confirmation: boolean
}

export interface IInvoiceReturnMaterials {
  materialID: number
  materialName: string
  unit: string
  holderAmount: number
  amount: number
  hasSerialNumber: boolean
  serialNumbers: string[]
  isDefective: boolean
  notes: string
}

export interface IInvoiceReturnView {
  id: number
  projectID: number
  districtName: string
  deliveryCode: string
  teamNumber: string
  teamLeaderName: string
  objectName: string
  objectSupervisorNames: string[]
  objectType: string
  acceptorName: string
  dateOfInvoice: Date
  confirmation: boolean
}
