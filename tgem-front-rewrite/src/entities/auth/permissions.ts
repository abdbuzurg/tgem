import IApiResponseFormat from "@shared/api/envelope";
import { USER_PATH } from "@shared/api/paths";
import axiosClient from "@shared/api/client";

// Effective permissions (v2). Each row is a (project, resource, action) the
// user is allowed. projectId === null means a global grant (applies in every
// project). The frontend declarative gate (Require) and the useAuth().can()
// hook consume this list.
export interface EffectivePermission {
  projectId: number | null
  resourceType: string
  action: string
}

export async function getEffectivePermissions(): Promise<EffectivePermission[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<EffectivePermission[]>>(
    `/${USER_PATH}/effective-permissions`
  )
  const responseData = responseRaw.data
  if (responseData.success) {
    return responseData.data
  } else {
    throw new Error(responseData.error)
  }
}
