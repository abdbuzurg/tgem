import { Link } from "react-router-dom"
import { REFERENCE_BOOK_DISTRICT, REFERENCE_BOOK_KL04KV_OBJECT, REFERENCE_BOOK_MATERIAL, REFERENCE_BOOK_MATERIAL_COST, REFERENCE_BOOK_MJD_OBJECT, REFERENCE_BOOK_OBJECTS, REFERENCE_BOOK_OPERATIONS, REFERENCE_BOOK_SIP_OBJECT, REFERENCE_BOOK_STVT_OBJECT, REFERENCE_BOOK_SUBSTATION_CELL_OBJECT, REFERENCE_BOOK_SUBSTATION_OBJECT, REFERENCE_BOOK_TEAM, REFERENCE_BOOK_TP_OBJECT, REFERENCE_BOOK_WORKER } from "@routes/paths"

const referenceBooksCategorized = [
  {
    category: "Объект",
    pages: [
      { name: "КЛ 04 КВ", url: REFERENCE_BOOK_KL04KV_OBJECT },
      { name: "МЖД", url: REFERENCE_BOOK_MJD_OBJECT },
      { name: "СИП", url: REFERENCE_BOOK_SIP_OBJECT },
      { name: "СТВТ", url: REFERENCE_BOOK_STVT_OBJECT },
      { name: "ТП", url: REFERENCE_BOOK_TP_OBJECT },
      { name: "Подстанция", url: REFERENCE_BOOK_SUBSTATION_OBJECT },
      { name: "Ячейки Подстанции", url: REFERENCE_BOOK_SUBSTATION_CELL_OBJECT },
      { name: "Другое", url: REFERENCE_BOOK_OBJECTS },
    ]
  },
  {
    category: "Сотрудники",
    pages: [
      { name: "Рабочий Персонал", url: REFERENCE_BOOK_WORKER },
      { name: "Бригады", url: REFERENCE_BOOK_TEAM },
    ]
  },
  {
    category: "Материалы",
    pages: [
      { name: "Материалы", url: REFERENCE_BOOK_MATERIAL },
      { name: "Ценники Материалов", url: REFERENCE_BOOK_MATERIAL_COST },
      // { name: "Поставщики", url: REFERENCE_BOOK_WORKER },
      { name: "Услуги", url: REFERENCE_BOOK_OPERATIONS },
    ]
  },
  {
    category: "Другое",
    pages: [
      { name: "Районы", url: REFERENCE_BOOK_DISTRICT },
    ]
  }
]

export default function ReferenceBooks() {
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
                className="bg-gray-800 text-white text-xl py-3 px-6 rounded cursor-pointer hover:bg-gray-900 text-center align-middle"
              >
                {page.name}
              </Link>
            )}
          </div>
        </div>
      ))}
    </main>
  )
}
