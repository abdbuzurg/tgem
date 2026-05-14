export interface IInvoiceOutputInProject {
  id: number
  projectID: number
  warehouseManagerWorkerID: number
  releasedWorkerID: number
  districtID: number
  teamID: number
  recipientWorkerID: number
  deliveryCode: string
  notes: string
  dateOfInvoice: Date
  confirmation: boolean
}

export interface IInvoiceOutputMaterials {
  materialID: number
  materialName: string
  unit: string
  warehouseAmount: number
  amount: number
  notes: string
  hasSerialNumber: boolean
  serialNumbers: string[]
}

export interface IInvoiceOutputInProjectView extends IInvoiceOutputInProject {
  warehouseManagerName: string
  releasedName: string
  districtName: string
  teamName: string
  recipientName: string
}
