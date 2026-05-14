export interface Permission {
  id: number
  roleID: number
  resourceID: number
  r: boolean
  w: boolean
  u: boolean
  d: boolean
}
