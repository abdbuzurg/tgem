export interface InvoiceMaterial {
  id: number
  materialCostID: number
  invoiceID: number
  invoiceType: "input" | "output" | "return" | "writeoff" | "object"
  amount: number
  notes: string
}

export interface InvoiceMaterialViewWithoutSerialNumbers {
  id: number
  materialName: string
  materialUnit: string
  isDefected: boolean
  costM19: number
  amount: number
  notes: string
}

export interface InvoiceMaterialViewWithSerialNumbers extends InvoiceMaterialViewWithoutSerialNumbers {
  serialNumbers: string[]
}


