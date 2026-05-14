import { Link, Outlet, useNavigate } from "react-router-dom"
import Button from "@shared/ui/Button"
import toast, { Toaster } from "react-hot-toast"
import { useQuery } from "@tanstack/react-query"
import { HOME, LOGIN, REFERENCE_BOOK, REPORT, STATISTICS } from "@routes/paths"
import { getProjectName } from "@entities/project/api"

export default function AppLayout() {
  const navigate = useNavigate()

  const logout = () => {
    const loadingToast = toast.loading("Выход.....")
    localStorage.removeItem("token")
    localStorage.removeItem("username")
    toast.dismiss(loadingToast)
    toast.success("Операция успешна")
    navigate(LOGIN)
  }

  const projectNameQuery = useQuery<string, Error, string>({
    queryKey: ["project-name"],
    queryFn: getProjectName,
  })

  return (
    <>
      <nav className="relative flex md:flex-row w-full justify-normal md:justify-between md:items-center bg-gray-800 px-3 py-2 shadow-lg text-gray-400">
        <div className="hidden md:block md:items-center md:justify-between">
          <ul className="flex p-0 font-medium space-x-8 items-center">
            <li>
              <Link to={`${HOME}`} className="block text-white bg-transparent p-0 hover:text-gray-400">
                Главная
              </Link>
            </li>
            <li>
              <Link to={`${REPORT}`} className="block text-white bg-transparent p-0 hover:text-gray-400">
                Отчет
              </Link>
            </li>
            <li>
              <Link to={`${REFERENCE_BOOK}`} className="block text-white bg-transparent p-0 hover:text-gray-400">
                Справочник
              </Link>
            </li>
            <li>
              <Link to={`${STATISTICS}`} className="block text-white bg-transparent p-0 hover:text-gray-400">
                Статистика
              </Link>
            </li>
          </ul>
        </div>
        {projectNameQuery.isSuccess && projectNameQuery.data &&
          <div className="flex flex-col text-white">
            <span className="font-bold italic">{projectNameQuery.data}</span>
          </div>
        }
        <div className="w-full flex items-center md:w-auto justify-between md:justify-normal space-x-4 font-medium ">
          <p className="text-4xl font-bold">ТГЭМ</p>
          <Button onClick={logout} text="Выход" />
        </div>
      </nav>
      <Toaster />
      <Outlet />
    </>
  )
}
