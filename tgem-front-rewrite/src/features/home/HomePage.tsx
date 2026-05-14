import { Link } from "react-router-dom";
import { HR_ATTENDANCE, INVOICE_CORRECTION, INVOICE_INPUT, INVOICE_OBJECT_PAGINATED_PAGE, INVOICE_OBJECT_USER, INVOICE_OUTPUT_IN_PROJECT, INVOICE_OUTPUT_OUT_OF_PROJECT, INVOICE_RETURN_OBJECT, INVOICE_RETURN_TEAM, LOSS_OBJECT, LOSS_TEAM, LOSS_WAREHOUSE, WRITEOFF_OBJECT, WRITEOFF_WAREHOUSE, } from "@routes/paths";
import MainReports from "@features/home/MainReports";

const homePageCategorized = [
  {
    category: "Приход",
    pages: [
      { name: "Приход", url: INVOICE_INPUT },
    ]
  },
  {
    category: "Отпуск",
    pages: [
      { name: "Отпуск внутри проекта", url: INVOICE_OUTPUT_IN_PROJECT },
      { name: "Отпуск вне проекта", url: INVOICE_OUTPUT_OUT_OF_PROJECT },
    ]
  },
  {
    category: "Возврат",
    pages: [
      { name: "Возврат c Бригады", url: INVOICE_RETURN_TEAM },
      { name: "Возврат c Объекта", url: INVOICE_RETURN_OBJECT },
    ]
  },
  {
    category: "Объект",
    pages: [
      { name: "Расходы Проекта", url: INVOICE_OBJECT_PAGINATED_PAGE },
      { name: "Расход материала на объект", url: INVOICE_OBJECT_USER },
      { name: "Корректировка", url: INVOICE_CORRECTION },
    ]
  },
  {
    category: "Списание",
    pages: [
      { name: "Акт списания со склада", url: WRITEOFF_WAREHOUSE },
      { name: "Акт утери со склада", url: LOSS_WAREHOUSE },
      { name: "Акт утери бригады", url: LOSS_TEAM },
      { name: "Акт утери с объекта", url: LOSS_OBJECT },
      { name: "Акт списания с объекта", url: WRITEOFF_OBJECT },
    ]
  },
  {
    category: "Отдел Кадров",
    pages: [
      { name: "МОРФО", url: HR_ATTENDANCE }
    ],
  },
]

export default function Home() {
  return (
    <div className="flex flex-col space-y-2">
      <div className="px-2">
        {homePageCategorized.map((book, index) => (
          <div className="py-1 px-2" key={index}>
            <span className="font-bold text-2xl">{book.category}</span>
            <div className="grid grid-cols-4 gap-2 px-2 py-2">
              {book.pages.map((page, index) =>
                <Link
                  key={index}
                  to={page.url}
                  className="bg-gray-800 text-white text-xl py-3 px-6 rounded cursor-pointer hover:bg-gray-900 text-center align-middle"
                >
                  {page.name}
                </Link>
              )}
            </div>
          </div>
        ))}
        <MainReports />
      </div>
    </div>
  )
}
