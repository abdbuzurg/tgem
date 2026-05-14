export default interface Material {
  id: number
  category: string
  code: string
  name: string
  unit: string
  notes: string
  article: string
  hasSerialNumber: boolean
  plannedAmountForProject: number
  showPlannedAmountInReport: boolean
}
