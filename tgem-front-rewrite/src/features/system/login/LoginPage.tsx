import { useEffect, useState } from "react"
import Input from "@shared/ui/Input"
import Button from "@shared/ui/Button"
import { useMutation, useQuery } from "@tanstack/react-query"
import loginUser from "@entities/auth/login"
import { getEffectivePermissions } from "@entities/auth/permissions"
import { useNavigate } from "react-router-dom"
import useAuth from "@app/hooks/useAuth"
import toast, { Toaster } from "react-hot-toast"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import Select from 'react-select'
import Project from "@entities/project/types"
import { GetAllProjects } from "@entities/project/api"
import { ADMINISTRATOR_HOME_PAGE, AUCTION_PRIVATE, HOME } from "@routes/paths"

export default function Login() {
  const navigate = useNavigate()
  const authContext = useAuth()

  // Clear any stale auth state when the login page first mounts. Intentionally
  // mount-only — depending on `authContext` re-fired this on every state
  // change (including the very setEffectivePermissions we make below on
  // login success), wiping the perms back to []. clearContext is now stable
  // via useCallback so we could include it as a dep, but mount-only keeps
  // the intent obvious.
  useEffect(() => {
    authContext.clearContext()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const [loginData, setLoginData] = useState({
    username: "",
    password: "",
    projectID: 0,
  })

  const [selectedProject, setSelectedProject] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [allProjects, setAllProjects] = useState<IReactSelectOptions<number>[]>([])
  const projectQuery = useQuery<Project[], Error, Project[]>({
    queryKey: ["all-projects"],
    queryFn: GetAllProjects,
  })
  useEffect(() => {
    if (projectQuery.isSuccess && projectQuery.data) {
      setAllProjects(projectQuery.data.map<IReactSelectOptions<number>>((value) => ({ label: value.name, value: value.id })))
    }
  }, [projectQuery.data, projectQuery.isSuccess])
  const onSelectedProject = (value: null | IReactSelectOptions<number>) => {
    if (!value) {
      setSelectedProject({ label: "", value: 0 })
      setLoginData({ ...loginData, projectID: 0 })
      return
    }

    setSelectedProject(value)
    setLoginData({ ...loginData, projectID: value.value })
  }

  const loginMutation = useMutation({ mutationFn: loginUser })
  const login = () => {

    if (loginData.username == "") {
      toast.error("Не указано имя пользователя")
      return
    }

    if (loginData.password == "") {
      toast.error("Не указан пароль")
      return
    }

    if (loginData.projectID == 0) {
      toast.error("Не выбран проект")
      return
    }

    const loadingToast = toast.loading("Выполняется вход.....")
    loginMutation.mutate(loginData, {
      onSuccess: (data) => {

        localStorage.setItem("token", data.token.toString())
        localStorage.setItem("username", loginData.username)

        authContext.setUsername(loginData.username)

        toast.dismiss(loadingToast)
        const successToast = toast.success("Вход прошел успешно.")

        // browser does not save token localStorage immediately
        // so settimeout is required
        setTimeout(async () => {
          try {
            const perms = await getEffectivePermissions()
            authContext.setEffectivePermissions(perms)
          } catch {
            // permissions fetch failure — proceed; AuthProvider will retry on next mount
          }

          switch (selectedProject.label) {
            case "Администрирование":
              if (data.admin) navigate(ADMINISTRATOR_HOME_PAGE)
              else toast.error("Система не распознает вас как администратора")
              break

            case "Auction":
              navigate(AUCTION_PRIVATE)
              break

            default:
              navigate(HOME)
              break
          }
          toast.dismiss(successToast)
        }, 1500)

      },
      onSettled: () => {
        toast.dismiss(loadingToast)
      }
    })
  }

  return (
    <div className="h-screen w-screen bg-gray-800">
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-1/2 flex flex-col items-center justify-evenly">
        <div className="bg-white p-4 rounded text-gray-800">
          <p className="font-bold text-4xl mb-4 text-center">ТГЭМ</p>
          <div className="basis-1 w-[350px]">
            <div className="flex flex-col justify-start mb-4">
              <label className="inline-block font-bold text-lg mb-2">Имя пользователя</label>
              <Input
                id="login_username"
                type="text"
                name="username"
                value={loginData.username}
                onChange={(e) => setLoginData({ ...loginData, [e.target.name]: e.target.value })}
              />
            </div>
            <div className="flex flex-col justify-start mb-3">
              <label className="inline-block font-bold text-lg mb-2">Пароль</label>
              <Input
                id="login_password"
                type="password"
                name="password"
                value={loginData.password}
                onChange={(e) => setLoginData({ ...loginData, [e.target.name]: e.target.value })}
              />
            </div>
            <div className="flex flex-col justify-start mb-3">
              <label className="inline-block font-bold text-lg mb-2">Проект</label>
              <div>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"materials"}
                  placeholder={""}
                  value={selectedProject}
                  options={allProjects}
                  onChange={(value) => onSelectedProject(value)}
                />
              </div>
            </div>
            <div className="flex justify-center">
              <Button onClick={login} text="Войти" />
            </div>
          </div>
        </div>
      </div>
      <Toaster />
    </div>
  )
}
