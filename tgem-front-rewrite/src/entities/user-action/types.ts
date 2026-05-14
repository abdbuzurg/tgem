export interface UserActionView {
  id: number
  actionURL: string
  actionType: string
  actionID: number
  actionStatus: boolean
  actionStatusMessage: string
  httpMethod: string
  requestIP: string
  userID: number
  projectID: number
  username: string
  dateOfAction: string
}

export interface UserActionPaginated {
  data: UserActionView[]
  count: number
  page: number
}

export interface UserActionFilter {
  userID?: number
  projectID?: number
  actionType?: string
  httpMethod?: string
  status?: boolean
  dateFrom?: string
  dateTo?: string
}

export interface UserActionFilterUserOption {
  id: number
  username: string
  workerName: string
}
