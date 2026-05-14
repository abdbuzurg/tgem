import { Link, Outlet, useNavigate } from "react-router-dom"
import Button from "@shared/ui/Button"
import { Toaster } from "react-hot-toast"
import toast from "react-hot-toast/headless"
import { ADMINISTRATOR_HOME_PAGE, LOGIN } from "@routes/paths"

export default function AdminLayout() {
  const navigate = useNavigate()

  const logout = () => {
    const loadingToast = toast.loading("Выход.....")
    localStorage.removeItem("token")
    localStorage.removeItem("username")
    toast.dismiss(loadingToast)
    toast.success("Операция успешна")
    navigate(LOGIN)
  }

  return (
    <>
      <nav className="relative flex md:flex-row w-full justify-normal md:justify-between md:items-center bg-gray-800 px-3 py-2 shadow-lg text-gray-400">
        <div className="hidden md:block md:items-center md:justify-between md:w-full">
          <ul className="flex p-0 font-medium space-x-8 items-center">
            <li>
              <Link to={`${ADMINISTRATOR_HOME_PAGE}`} className="block text-white bg-transparent p-0 hover:text-gray-400">
                Главная
              </Link>
            </li>
          </ul>
        </div>
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
