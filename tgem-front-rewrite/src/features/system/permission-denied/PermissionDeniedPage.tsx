import { Link } from "react-router-dom";

export default function PermissionDenied() {
  return (
    <div id='error-page' className='flex flex-col gap-8 justify-center items-center h-screen'>
      <h1 className='text-4xl font-bold'>403</h1>
      <p>Доступ запрещен</p>
      <Link  to={"/home"} className="bg-gray-800 text-white text-2xl py-2 px-4 rounded cursor-pointer hover:bg-gray-900">На главную</Link>
    </div>
  )
}
