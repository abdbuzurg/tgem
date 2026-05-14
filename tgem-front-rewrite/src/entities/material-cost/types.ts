export interface IMaterialCost{
  id: number
  materialID: number
  costPrime: number
  costM19: number
  costWithCustomer: number
}

export interface IMaterialCostView extends Omit<IMaterialCost, "materialID"> {
  materialName: string
}
