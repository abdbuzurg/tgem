import { useInfiniteQuery } from "@tanstack/react-query"
import { InvoiceCorrectionPaginated, InvoiceCorrectionPaginatedView, InvoiceCorrectionSearchParameters, getPaginatedInvoiceCorrection } from "./api"
import { useEffect, useState } from "react"
import LoadingDots from "@shared/ui/LoadingDots"
import Button from "@shared/ui/Button"
import CorrectionModal from "./CorrectionModal"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import { objectTypeIntoRus } from "@shared/lib/data/objectStatuses"
import SearchInvoiceCorrection from "./SearchInvoiceCorrection"
import useScrollPaginated from "@shared/hooks/useScrollPaginated"

export default function InvoiceCorrection() {
  //SearchModal Logic
  const [showSearchModal, setShowSearchModal] = useState(false)
  const [searchParameters, setSearchParameters] = useState<InvoiceCorrectionSearchParameters>({
    objectID: 0,
    teamID: 0,
  })

  //FETCHING LOGIC
  const tableDataQuery = useInfiniteQuery<InvoiceCorrectionPaginated, Error>({
    queryKey: ["invoice-correction", searchParameters],
    queryFn: ({ pageParam }) => getPaginatedInvoiceCorrection({ pageParam }, searchParameters),
    getNextPageParam: (lastPage) => {
      if (lastPage.page * ENTRY_LIMIT > lastPage.count) return undefined
      return lastPage.page + 1
    }
  })
  const [tableData, setTableData] = useState<InvoiceCorrectionPaginatedView[]>([])
  useEffect(() => {
    if (tableDataQuery.isSuccess && tableDataQuery.data) {
      const data: InvoiceCorrectionPaginatedView[] = tableDataQuery.data.pages.reduce<InvoiceCorrectionPaginatedView[]>((acc, page) => [...acc, ...page.data], [])
      setTableData(data)
    }
  }, [tableDataQuery.data, tableDataQuery.isSuccess])

  useScrollPaginated(tableDataQuery.fetchNextPage)

  const [showCorrectionModal, setShowCorrectionModal] = useState(false)
  const [modalData, setModalData] = useState<InvoiceCorrectionPaginatedView>()
  const toggleModal = (val: InvoiceCorrectionPaginatedView) => {
    setModalData(val)
    setShowCorrectionModal(true)
  }

  return (
    <main>
      <div className="mt-2 px-2 flex gap-x-2">
        <span className="font-bold text-3xl">Корректировка поступления</span>
        <Button onClick={() => setShowSearchModal(true)} text="Поиск" buttonType="default" />
        <Button onClick={() => setSearchParameters({
          teamID: 0,
          objectID: 0,
        })} text="Сброс поиска" buttonType="delete" />
      </div>
      <table className="table-auto text-sm text-left mt-2 w-full border-box">
        <thead className="shadow-md border-t-2">
          <tr>
            <th className="px-4 py-3 w-[100px]">
              <span>Код</span>
            </th>
            <th className="px-4 py-3">
              <span>Супервайзер</span>
            </th>
            <th className="px-4 py-3">
              <span>Объект</span>
            </th>
            <th className="px-4 py-3 w-[150px]">
              <span>Бригада</span>
            </th>
            <th className="px-4 py-3 w-[150px]">
              <span>Дата</span>
            </th>
            <th className="px-4 py-3"></th>
          </tr>
        </thead>
        <tbody>
          {tableDataQuery.isLoading &&
            <tr>
              <td colSpan={6}>
                <LoadingDots />
              </td>
            </tr>
          }
          {tableDataQuery.isError &&
            <tr>
              <td colSpan={6} className="text-red font-bold text-center">
                {tableDataQuery.error.message}
              </td>
            </tr>
          }
          {tableDataQuery.isSuccess &&
            tableData.map((row, index) => (
              <tr key={index} className="border-b">
                <td className="px-4 py-3">{row.deliveryCode}</td>
                <td className="px-4 py-3">{row.supervisorName}</td>
                <td className="px-4 py-3">{row.objectName} ({objectTypeIntoRus(row.objectType)})</td>
                <td className="px-4 py-3">{row.teamLeaderName}</td>
                <td className="px-4 py-3">{row.dateOfInvoice.toString().substring(0, 10)}</td>
                <td className="px-4 py-3 border-box flex space-x-3">
                  <Button onClick={() => toggleModal(row)} text="Корректировать" />
                </td>
              </tr>
            ))
          }
          {tableDataQuery.hasNextPage &&
            <tr>
              <td colSpan={7}>
                <div className="w-full py-4 flex justify-center">
                  <div
                    onClick={() => tableDataQuery.fetchNextPage()}
                    className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
                  >
                    {tableDataQuery.isLoading && <LoadingDots height={30} />}
                    {!tableDataQuery.isLoading && "Загрузить еще"}
                  </div>
                </div>
              </td>
            </tr>
          }
        </tbody>
      </table>
      {showCorrectionModal &&
        <CorrectionModal
          setShowModal={setShowCorrectionModal}
          invoiceObject={modalData!}
        />
      }
      {showSearchModal && <SearchInvoiceCorrection
        searchParameters={searchParameters}
        setSearchParameters={setSearchParameters}
        setShowModal={setShowSearchModal}
      />
      }
    </main>
  )
}
