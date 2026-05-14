import { Link } from "react-router-dom"
import { ADMINISTRATOR_PROJECT, ADMINISTRATOR_USERS, ADMINISTRATOR_USER_ACTIONS, ADMINISTRATOR_WORKERS } from "@routes/paths"

const referenceBooksCategorized = [
  {
    category: "Проекты",
    pages: [
      { name: "Проекты", url: ADMINISTRATOR_PROJECT },
      { name: "Пользователи", url: ADMINISTRATOR_USERS },
      { name: "Детализация Пользователей", url: ADMINISTRATOR_WORKERS },
      { name: "Журнал действий пользователей", url: ADMINISTRATOR_USER_ACTIONS },
    ]
  },
]
export default function AdministatorHome() {
  return (
    <main>
      {referenceBooksCategorized.map((book, index) => (
        <div className="py-1 px-2" key={index}>
          <span className="font-bold text-2xl">{book.category}</span>
          <div className="grid grid-cols-4 gap-2 px-2 py-2">
            {book.pages.map((page, index) =>
              <Link
                key={index}
                to={page.url}
                className="flex items-center text-center bg-gray-800 text-white text-xl py-3 px-6 rounded cursor-pointer hover:bg-gray-900 text-center align-middle"
              >
                <span className="w-full">{page.name}</span>
              </Link>
            )}
          </div>
        </div>
      ))}
    </main>
  )
}
