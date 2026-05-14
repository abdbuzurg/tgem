import { useQuery } from "@tanstack/react-query"
import Modal from "@shared/components/Modal"
import { UserPermission, getPermissionsByRoleName } from "@entities/permission/api"
import { useState, useEffect, Fragment } from "react"
import { IoIosCheckmarkCircle } from "react-icons/io";
import { MdCancel } from "react-icons/md";

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
  workerName: string
  roleName: string
}

export default function UserPermissionModal({
  setShowModal,
  workerName,
  roleName,
}: Props) {

  //Permission Get Logic
  const [permissions, setPermissions] = useState<UserPermission[]>([])
  const userPermissionQuery = useQuery<UserPermission[], Error, UserPermission[]>({
    queryKey: [`permission-by-role-name-${roleName}`],
    queryFn: () => getPermissionsByRoleName(roleName)
  })
  useEffect(() => {

    if (userPermissionQuery.isSuccess && userPermissionQuery.data) {
      setPermissions(userPermissionQuery.data)
    }

  }, [userPermissionQuery.data, userPermissionQuery.isSuccess])


  return (
    <Modal setShowModal={setShowModal} bigModal>
      <div className="flex flex-col space-y-2">
        <span className="font-bold text-2xl">Уровень доступа пользователя - {workerName}</span>
        <div className="grid grid-cols-5 gap-y-2 items-center text-center max-h-[50vh] overflow-y-scroll">
          {/* Headers of the table */}
          <div className="font-semibold">Имя ресурса</div>
          <div className="font-semibold">Просмотр</div>
          <div className="font-semibold">Добавление</div>
          <div className="font-semibold">Изменение</div>
          <div className="font-semibold">Удаление</div>

          {/* Content of the table */}
          {permissions.map((permission, index) => (
            <Fragment key={index}>
              <div>{permission.resourceName}</div>
              <div className="flex justify-center">
                {permission.r ? <IoIosCheckmarkCircle className="text-green-500" /> : <MdCancel className="text-red-500" />}
              </div>
              <div className="flex justify-center">
                {permission.w ? <IoIosCheckmarkCircle className="text-green-500" /> : <MdCancel className="text-red-500" />}
              </div>
              <div className="flex justify-center">
                {permission.u ? <IoIosCheckmarkCircle className="text-green-500" /> : <MdCancel className="text-red-500" />}
              </div>
              <div className="flex justify-center">
                {permission.d ? <IoIosCheckmarkCircle className="text-green-500" /> : <MdCancel className="text-red-500" />}
              </div>
            </Fragment>
          ))}
        </div>
      </div>
    </Modal>
  )
}
