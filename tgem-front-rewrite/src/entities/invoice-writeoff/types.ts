export interface IInvoiceWriteOff {
  id: number
  projectID: number
  releasedWorkerID: number
  writeOffType: string
  writeOffLocationID: number
  dateOfInvoice: Date
  confirmation: boolean
  dateOfConfirmation: Date
  deliveryCode: string
}

export interface IInvoiceWriteOffMaterials {
  materialID: number
  materialName: string
  unit: string
  amount: number
  notes: string
  hasSerialNumber: boolean
  serialNumbers: string[]
  locationAmount: number
}

export interface IInvoiceWriteOffView extends IInvoiceWriteOff  {
  writeOffLocationName: string
  releasedWorkerName: string
}
