export default interface Project {
  id: number
  name: string
  client: string
  budget: number
  budgetCurrency: string
  description: string
  signedDateOfContract: Date
  dateStart: Date
  dateEnd: Date
  projectManager: string
}
