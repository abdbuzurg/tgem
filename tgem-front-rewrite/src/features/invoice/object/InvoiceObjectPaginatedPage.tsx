import { useInfiniteQuery } from "@tanstack/react-query";
import { ENTRY_LIMIT } from "@shared/config/pagination";
import { useEffect, useState } from "react";
import LoadingDots from "@shared/ui/LoadingDots";
import "react-datepicker/dist/react-datepicker.css";
import { FaRegListAlt } from "react-icons/fa";
import IconButton from "@shared/components/IconButtons";
import { InvoiceObject, InvoiceObjectView, getInvoiceObject } from "@features/invoice/object/api";
import { objectTypeIntoRus } from "@shared/lib/data/objectStatuses";
import ShowInvoiceObjectDetails from "./ShowInvoiceObjectDetails";
import useScrollPaginated from "@shared/hooks/useScrollPaginated"

export default function InvoiceObjectPaginatedPage() {
  //FETCHING LOGIC
  const tableDataQuery = useInfiniteQuery<InvoiceObject, Error>({
    queryKey: ["invoice-objects-paginated"],
    queryFn: ({ pageParam }) => getInvoiceObject({ pageParam }),
    getNextPageParam: (lastPage) => {
      if (lastPage.page * ENTRY_LIMIT > lastPage.count) return undefined
      return lastPage.page + 1
    }
  })
  const [tableData, setTableData] = useState<InvoiceObjectView[]>([])
  useEffect(() => {
    if (tableDataQuery.isSuccess && tableDataQuery.data) {
      const data: InvoiceObjectView[] = tableDataQuery.data.pages.reduce<InvoiceObjectView[]>((acc, page) => [...acc, ...page.data], [])
      setTableData(data)
    }
  }, [tableDataQuery.data, tableDataQuery.isSuccess])

  useScrollPaginated(tableDataQuery.fetchNextPage)

  //Details logic
  const [showDetailsModal, setShowDetailsModal] = useState(false)
  const [detailModalData, setDetailModalData] = useState<InvoiceObjectView>({
    id: 0,
    deliveryCode: "",
    supervisorName: "",
    objectName: "",
    objectType: "",
    teamNumber: "",
    dateOfInvoice: new Date(),
    confirmedByOperator: false,
  })

  const showDetails = (index: number) => {
    setDetailModalData(tableData[index])
    setShowDetailsModal(true)
  }

  return (
    <main>
      <div className="mt-2 px-2 flex justify-between">
        <span className="text-3xl font-bold">Расходы проекта</span>
      </div>
      <table className="table-auto text-sm text-left mt-2 w-full border-box">
        <thead className="shadow-md border-t-2">
          <tr>
            <th className="px-4 py-3 w-[150px]">
              <span>Код</span>
            </th>
            <th className="px-4 py-3">
              <span>Объект</span>
            </th>
            <th className="px-4 py-3">
              <span>Бригада</span>
            </th>
            <th className="px-4 py-3 w-[150px]">
              <span>Дата</span>
            </th>
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
              <tr key={index} className="border-b hover:bg-gray-200">
                <td className="px-4 py-3">{row.deliveryCode}</td>
                <td className="px-4 py-3">{`${row.objectName}(${objectTypeIntoRus(row.objectType)})`}</td>
                <td className="px-4 py-3">{row.teamNumber}</td>
                <td className="px-4 py-3">{row.dateOfInvoice.toString().substring(0, 10)}</td>
                <td className="px-4 py-3 border-box flex space-x-3">
                  <IconButton
                    icon={<FaRegListAlt size="20px" title={`Просмотр деталей накладной ${row.deliveryCode}`} />}
                    onClick={() => showDetails(index)}
                  />
                </td>
              </tr>
            ))
          }
          {tableDataQuery.hasNextPage &&
            <tr>
              <td colSpan={5}>
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
      {showDetailsModal && <ShowInvoiceObjectDetails setShowModal={setShowDetailsModal} data={detailModalData} />}
    </main>
  )
}
