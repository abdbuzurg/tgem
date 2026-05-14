export interface InvoiceOutputOutOfProject {
  id: number
  projectID: number
  nameOfProject: string
  deliveryCode: string
  releasedWorkerID: number
  dateOfInvoice: Date
  confirmation: boolean
  notes: string
}
