import { Fragment, useEffect, useState } from "react"
import Modal from "@shared/components/Modal"
import Input from "@shared/ui/Input"
import { Accordion } from "flowbite-react"
import { Permission } from "@entities/permission/types"
import { IRole } from "@entities/role/types"
import Button from "@shared/ui/Button"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { createRole } from "@entities/role/api"
import { createPermissions } from "@entities/permission/api"
import toast from "react-hot-toast"
import Resource from "@entities/resource/types"
import { getAllResources } from "@entities/resource/api"

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
}

interface AccordionShape {
  category: string
  permissions: {
    id: number
    name: string
    url: string
  }[]
}

export default function AddNewRoleModal({ setShowModal }: Props) {

  const [permissions, setPermissions] = useState<Permission[]>([])

  //Resources Data Logic
  const [resources, setResources] = useState<AccordionShape[]>([])
  const resourcesQuery = useQuery<Resource[], Error, Resource[]>({
    queryKey: ["all-resources"],
    queryFn: getAllResources,
  })
  useEffect(() => {

    if (resourcesQuery.isSuccess && resourcesQuery.data) {

      const data: AccordionShape[] = []
      let index
      for (const one of resourcesQuery.data) {
        index = data.findIndex((val) => val.category == one.category)
        if (index == -1) {
          data.push({
            category: one.category,
            permissions: [
              {
                id: one.id,
                name: one.name,
                url: one.url
              }
            ]
          })
        } else {
          data[index].permissions.push({
            id: one.id,
            name: one.name,
            url: one.url,
          })
        }

      }

      setResources(data)
      setPermissions(resourcesQuery.data.map<Permission>((val) => ({
        id: 0,
        resourceID: val.id,
        roleID: 0,
        r: false,
        w: false,
        u: false,
        d: false,
      })))
    }
  }, [resourcesQuery.data, resourcesQuery.isSuccess])

  //Role Data
  const [roleData, setRoleData] = useState<IRole>({
    id: 0,
    name: "",
    description: "",
  })


  const onChange = (e: React.FormEvent<HTMLInputElement>, resourceID: number) => {

    const index = permissions.findIndex((val) => val.resourceID == resourceID)
    const value = e.currentTarget.value

    if (value == "r") permissions[index].r = e.currentTarget.checked 
    if (value == "w") permissions[index].w = e.currentTarget.checked 
    if (value == "u") permissions[index].u = e.currentTarget.checked 
    if (value == "d") permissions[index].d = e.currentTarget.checked 

    setPermissions(permissions)
  }

  const queryClient = useQueryClient()

  const createPermissionsMutation = useMutation<boolean, Error, Permission[]>({
    mutationFn: createPermissions,
  })

  const createRoleMutation = useMutation<IRole, Error, IRole>({
    mutationFn: createRole,
    onSuccess: (data: IRole) => {
      createPermissionsMutation.mutate(permissions.map((value) => ({...value, roleID: data.id})))
      queryClient.invalidateQueries(["all-roles"])
      setShowModal(false)
    }
  })

  const onSubmitRole = () => {

    if (roleData.name == "") {
      toast.error("Не указано имя роли")
      return
    }

    if (roleData.description == "") {
      toast.error("Не указано описание для роли")
      return
    }

    if (permissions.length == 0) {
      toast.error("Роль должна иметь хоть вид 1 доступ.")
      return
    }

    const loadingToast = toast.loading("Сохранение новых данных...")
    createRoleMutation.mutate(roleData, {
      onSuccess: () => {
        toast.success("Сохранение прошло успешно")
        setShowModal(false)
      },
      onSettled: () => {
        toast.dismiss(loadingToast)
      }
    })

  }

  return (
    <Modal setShowModal={setShowModal} bigModal>
      <div className="flex space-x-2 max-h-[70vh]">
        <div className="flex flex-col space-y-2 w-[350px]">
          <h3 className="text-2xl font-medium text-gray-800">
            Добавление роли
          </h3>
          <div className="flex flex-col space-y-1">
            <label htmlFor="roleName">Имя роли</label>
            <Input
              id="roleName"
              value={roleData.name}
              type="text"
              onChange={(e) => setRoleData({ ...roleData, name: e.target.value })}
              name={"roleName"}
            />
          </div>
          <div className="flex flex-col space-y-1">
            <label htmlFor="description">Описание роли</label>
            <Input
              id="description"
              value={roleData.description}
              type="text"
              onChange={(e) => setRoleData({ ...roleData, description: e.target.value })}
              name={"description"}
            />
          </div>
          <div>
            <Button
              text="Добавить"
              onClick={() => onSubmitRole()}
            />
          </div>
        </div>
        <div className="flex flex-col space-y-1 w-full">
          <span>Список доступов для роли</span>
          <Accordion>
            {resources.map((resource, index) => (
              <Accordion.Panel key={index} className="outline-none">
                <Accordion.Title className="h-5">
                  {resource.category[0].toUpperCase() + resource.category.slice(1)}
                </Accordion.Title>
                <Accordion.Content className="px-2 py-2 overflow-y-auto max-h-[40vh]">
                  <div className="w-full text-center grid grid-cols-5 auto-cols-max place-items-center gap-y-1 text-sm">
                    <div className="w-full px-2 py-1">
                      <span className="font-semibold">Ресурс</span>
                    </div>
                    <div className="w-full px-2 py-1">
                      <span className="font-semibold">Просмотр</span>
                    </div>
                    <div className="w-full px-2 py-1">
                      <span className="font-semibold">Добавление</span>
                    </div>
                    <div className="w-full px-2 py-1">
                      <span className="font-semibold">Изменение</span>
                    </div>
                    <div className="w-full px-2 py-1">
                      <span className="font-semibold">Удаление</span>
                    </div>
                    {resource.permissions.map((value, resourceIndex) => (
                      <Fragment key={resourceIndex}>
                        <div className="w-full px-2 py-1">
                          <span>
                            {value.name[0].toUpperCase() + value.name.slice(1)}
                          </span>
                        </div>
                        <div className="w-full px-2 py-1">
                          <input
                            type="checkbox"
                            name={value.url}
                            value="r"
                            onChange={(e) => onChange(e, value.id)}
                          />
                        </div>
                        <div className="w-full px-2 py-1">
                          <input
                            type="checkbox"
                            name={value.url}
                            value="w"
                            onChange={(e) => onChange(e, value.id)}
                          />
                        </div>
                        <div className="w-full px-2 py-1">
                          <input
                            type="checkbox"
                            name={value.url}
                            value="u"
                            onChange={(e) => onChange(e, value.id)}
                          />
                        </div>
                        <div className="w-full px-2 py-1">
                          <input
                            type="checkbox"
                            name={value.url}
                            value="d"
                            onChange={(e) => onChange(e, value.id)}
                          />
                        </div>
                      </Fragment>
                    ))}
                  </div>
                </Accordion.Content>
              </Accordion.Panel>
            ))}
          </Accordion>
        </div>
      </div>
    </Modal>
  )
}
