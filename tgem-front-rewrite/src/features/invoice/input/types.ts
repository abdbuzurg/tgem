export interface IInvoiceInput {
  id: number
  projectID: number
  warehouseManagerWorkerID: number
  releasedWorkerID: number
  deliveryCode: string
  notes: string
  dateOfInvoice: Date
  confirmation: boolean,
}

export interface IInvoiceInputView extends Omit<IInvoiceInput, "warehouseManagerWorkerID" | "releasedWorkerID"> {
  warehouseManagerName: string
  releasedName: string
}

export interface IInvoiceInputMaterials {
  materialID: number
  materialName: string
  unit: string
  amount: number
  materialCostID: number
  materialCost: number
  notes: string
  hasSerialNumber: boolean
  serialNumbers: string[]
}
