import toast from "react-hot-toast"
import { importInvoiceInput } from "@features/invoice/input/api"
import { importInvoiceOutput } from "@features/invoice/output-in-project/api"

export default function Import() {

  const acceptImportInvoiceInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return
    importInvoiceInput(e.target.files[0])
      .then(res => res ? toast.success("Импортирование успешно") : toast.error("Провал"))
      .catch(err => toast.error(`Провал: ${err}`))
  }

  const acceptImportInvoiceOutput = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return
    importInvoiceOutput(e.target.files[0])
      .then(res => res ? toast.success("Импортирование успешно") : toast.error("Провал"))
      .catch(err => toast.error(`Провал: ${err}`))
  }


  return (
    <main>
      <div className="flex space-x-2 py-2 px-3">
        <div className="bg-gray-800 text-white text-2xl py-3 px-6 rounded cursor-pointer hover:bg-gray-900">
          <label
            htmlFor="file-input"
            className="w-full text-white py-3 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer text-center"
          >
            Импорт Приход
          </label>

          <input
            name="file-input"
            type="file"
            id="file-input"
            onChange={(e) => acceptImportInvoiceInput(e)}
            className="hidden"
          />
        </div>
        <div className="bg-gray-800 text-white text-2xl py-3 px-6 rounded cursor-pointer hover:bg-gray-900">
          <label
            htmlFor="file-output"
            className="w-full text-white py-3 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer text-center"
          >
            Импорт Отпуск
          </label>

          <input
            name="file-output"
            type="file"
            id="file-output"
            onChange={(e) => acceptImportInvoiceOutput(e)}
            className="hidden"
          />
        </div>
      </div>
    </main>
  )
}
