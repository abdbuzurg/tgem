import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import IReactSelectOptions from "@shared/types/ReactSelectOptions";
import Modal from "@shared/components/Modal";
import Select, { MultiValue } from 'react-select'
import IWorker from "@entities/worker/types";
import { IRole } from "@entities/role/types";
import { useEffect, useState } from "react";
import Button from "@shared/ui/Button";
import Input from "@shared/ui/Input";
import { getAllRoles } from "@entities/role/api";
import AddNewWorkerModal from "./AddNewWorkerModal";
import AddNewRoleModal from "./AddNewRoleModal";
import toast from "react-hot-toast";
import { NewUserData, UserView, createUser, updateUser } from "@entities/user/api";
import { GetAllProjects } from "@entities/project/api";
import Project from "@entities/project/types";
import { getAllWorkers } from "@entities/worker/api";
import LoadingDots from "@shared/ui/LoadingDots";

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
  userData: UserView
}

export default function MutationUserModal({
  setShowModal,
  userData,
}: Props) {

  //Worker Select logic
  const [selectedWorkerID, setSelectedWorkerID] = useState<IReactSelectOptions<number>>({
    value: 0,
    label: ""
  })
  const [availableWorkers, setAvailableWorkers] = useState<IReactSelectOptions<number>[]>([])
  const allWorkersQuery = useQuery<IWorker[], Error, IWorker[]>({
    queryKey: ["all-workers"],
    queryFn: getAllWorkers,
  })
  useEffect(() => {
    if (allWorkersQuery.isSuccess && allWorkersQuery.data) {
      const workers = allWorkersQuery.data.map<IReactSelectOptions<number>>(value => ({
        label: value.name,
        value: value.id,
      }))

      setAvailableWorkers(workers)
      if (userData.workerName != "") {
        const curWorker = workers.find(v => v.label == userData.workerName)
        if (curWorker) {
          setSelectedWorkerID(curWorker)
        }
      }
    }
  }, [allWorkersQuery.data, allWorkersQuery.isSuccess, userData.workerName])
  // Adding new worker modal logic
  const [showAddNewWorkerModal, setShowAddNewWorkerModal] = useState(false)

  //Username Logic
  const [username, setUsername] = useState(userData.username)

  //Password Logic
  const [password, setPassword] = useState("")

  //Select Role logic
  const [selectedRoleID, setSelectedRoleID] = useState<IReactSelectOptions<number>>({
    label: "",
    value: 0,
  })
  const [availableRoles, setAvailableRoles] = useState<IReactSelectOptions<number>[]>([])
  const allRolesQuery = useQuery<IRole[], Error, IRole[]>({
    queryKey: ["all-roles"],
    queryFn: getAllRoles,
  })

  useEffect(() => {
    if (allRolesQuery.isSuccess && allRolesQuery.data) {
      const roles = allRolesQuery.data.map<IReactSelectOptions<number>>((value) => ({
        value: value.id,
        label: value.name,
      }))

      setAvailableRoles(roles)

      if (userData.roleName != "") {
        const curRoles = roles.find(v => v.label == userData.roleName)
        if (curRoles) {
          setSelectedRoleID(curRoles)
        }
      }
    }

  }, [allRolesQuery.data, allRolesQuery.isSuccess, userData.roleName])

  // Adding new role logic
  const [showAddNewRoleModal, setShowAddNewRoleModal] = useState(false)

  //Selecting projects logic
  const [selectedProjects, setSelectedProjects] = useState<IReactSelectOptions<number>[]>([])
  const [availableProjects, setAvailableProjects] = useState<IReactSelectOptions<number>[]>([])
  const allProjectsQuery = useQuery<Project[], Error, Project[]>({
    queryKey: ["all-projects"],
    queryFn: GetAllProjects,
  })

  useEffect(() => {
    if (allProjectsQuery.isSuccess && allProjectsQuery.data) {
      const projects = allProjectsQuery.data.map<IReactSelectOptions<number>>((value) => ({
        value: value.id,
        label: value.name,
      }))

      setAvailableProjects(projects)

      if (userData.accessToProjects) {
        const userInPorjects: IReactSelectOptions<number>[] = []
        userData.accessToProjects.map(v => {
          const curProject = projects.find(proj => proj.label == v)
          if (curProject) {
            userInPorjects.push(curProject)
          }
        })
        setSelectedProjects(userInPorjects)
      }
    }
  }, [allProjectsQuery.data, allProjectsQuery.isSuccess, userData.accessToProjects])

  const onProjectSelect = (value: MultiValue<IReactSelectOptions<number>>) => {
    setSelectedProjects([...value])
  }

  // Submiting new user Data
  const queryClient = useQueryClient()

  const createUserMutation = useMutation<boolean, Error, NewUserData>({
    mutationFn: createUser,
  })

  const updateUserMutation = useMutation<boolean, Error, NewUserData>({
    mutationFn: updateUser,
  })

  const onSubmitUser = () => {

    if (selectedWorkerID.value == 0) {
      toast.error("Сотрудник не выбран")
      return
    }

    if (username == "") {
      toast.error("Лоин пользователя не был указан")
      return
    }

    if (password == "" && userData.id == 0) {
      toast.error("Пароль не был указан")
      return
    }

    if (selectedRoleID.value == 0) {
      toast.error("Не выбрана роль для пользователя")
      return
    }

    if (selectedProjects.length == 0) {
      toast.error("Не выбран проект(-ы) для пользователя")
      return
    }

    const loadingToast = toast.loading("Сохранение новых данных...")
    if (userData.id == 0) {
      createUserMutation.mutate({
        userData: {
          id: 0,
          workerID: selectedWorkerID.value,
          roleID: selectedRoleID.value,
          username: username,
          password: password,
        },
        projects: [...selectedProjects.map<number>(value => value.value)],
      }, {
        onSuccess: () => {
          queryClient.invalidateQueries(["users"])
          toast.success("Сохранение данных прошло успешно")
          setShowModal(false)
        },
        onSettled: () => {
          toast.dismiss(loadingToast)
        }
      })

      return
    }

    updateUserMutation.mutate({
      userData: {
        id: userData.id,
        workerID: selectedWorkerID.value,
        roleID: selectedRoleID.value,
        username: username,
        password: password,
      },
      projects: [...selectedProjects.map<number>(value => value.value)]
    }, {
      onSuccess: () => {
        queryClient.invalidateQueries(["users"])
        toast.success("Сохранение данных прошло успешно")
        setShowModal(false)
      },
      onSettled: () => {
        toast.dismiss(loadingToast)
      }
    })
  }

  return (
    <Modal setShowModal={setShowModal} >
      <div className="flex flex-col space-y-2">
        <div>
          <span className="text-2xl font-bold">Добавление нового пользователя</span>
        </div>
        <div className="flex flex-col space-y-3">
          <span className="text-xl font-semibold">Информация пользователя</span>
          <label htmlFor="worker">Выберите соотрудника</label>
          <div className="flex space-x-2 align-middle items-center">
            <div className="flex flex-col space-y-1">
              <div id="worker" className="w-[250px]">
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"teams"}
                  placeholder={""}
                  value={selectedWorkerID}
                  options={availableWorkers}
                  onChange={(value) => setSelectedWorkerID({
                    value: value?.value ?? 0,
                    label: value?.label ?? "",
                  })}
                />
              </div>
            </div>
            <Button
              text="Добавить сотрудрика"
              onClick={() => setShowAddNewWorkerModal(true)}
            />
          </div>
          <div className="flex space-x-2 items-center">
            <div className="flex flex-col space-y-1">
              <label htmlFor="username">Логин пользователя</label>
              <Input
                id="username"
                value={username}
                name="username"
                type="text"
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setUsername(e.target.value)}
              />
            </div>
            <div className="flex flex-col space-y-1">
              <label htmlFor="password">Пароль</label>
              <Input
                id="password"
                value={password}
                name="password"
                type="password"
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setPassword(e.target.value)}
              />
            </div>
          </div>
          {userData.id != 0 &&
            <span className="italic text-sm text-gray-500">Пароль пользователя останется таким же если оставить пустым</span>
          }
        </div>
        <div className="flex flex-col space-y-2">
          <span className="text-xl font-semibold">Доступы пользователя</span>
          <div className="flex space-x-2 align-middle items-center">
            <div className="flex flex-col space-y-1">
              <div id="worker" className="w-[250px]">
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"teams"}
                  placeholder={""}
                  value={selectedRoleID}
                  options={availableRoles}
                  onChange={(value) => setSelectedRoleID({
                    value: value?.value ?? 0,
                    label: value?.label ?? "",
                  })}
                />
              </div>
            </div>
            <Button
              text="Добавить роль"
              onClick={() => setShowAddNewRoleModal(true)}
            />
          </div>
        </div>
        <div className="flex flex-col space-y-2">
          <span className="text-xl font-semibold">Доступы в проекты для пользователя</span>
          <div className="flex space-x-2 align-middle items-center">
            <div className="flex flex-col space-y-1">
              <div id="projects" className="w-[300px]">
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  isMulti
                  name={"projects"}
                  placeholder={""}
                  value={selectedProjects}
                  options={availableProjects}
                  onChange={onProjectSelect}
                />
              </div>
            </div>
          </div>
        </div>
        <div>
          {createUserMutation.isLoading
            ?
            <LoadingDots height={30} />
            :
            <Button
              text="Созать пользователя"
              onClick={() => onSubmitUser()}
            />
          }
        </div>
      </div>
      {showAddNewWorkerModal && <AddNewWorkerModal setShowModal={setShowAddNewWorkerModal} />}
      {showAddNewRoleModal && <AddNewRoleModal setShowModal={setShowAddNewRoleModal} />}
    </Modal>
  )
}
