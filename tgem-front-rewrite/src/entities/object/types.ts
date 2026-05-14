export interface IObject {
  id: number
  projectID: number
  objectDetailedID: number
  type: string
  name: string
  status: string
}

export interface SecondaryObjectData {
  model: string,
  amountStores: number,
  amountEntrances: number,
  hasBasement: boolean,
  voltageClass: string,
  nourashes: string,
  ttCoefficient: string,
  amountFeeders: number,
  length: number,
} 

export interface ObjectDataForSelect {
  id: number
  objectName: string
  objectType: string
}
