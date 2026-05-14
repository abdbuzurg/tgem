import { useNavigate } from "react-router-dom";
import Button from "@shared/ui/Button";
import { INVOICE_OBJECT_MUTATION_ADD } from "@routes/paths";
import { useInfiniteQuery } from "@tanstack/react-query";
import { InvoiceObject, InvoiceObjectView, getInvoiceObject } from "@features/invoice/object/api";
import { ENTRY_LIMIT } from "@shared/config/pagination";
import { useEffect, useState } from "react";
import LoadingDots from "@shared/ui/LoadingDots";
import { objectTypeIntoRus } from "@shared/lib/data/objectStatuses";
import useScrollPaginated from "@shared/hooks/useScrollPaginated"

export default function InvoiceObjectUser() {
  const tableDataQuery = useInfiniteQuery<InvoiceObject, Error>({
    queryKey: ["invoice-object-user"],
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


  const navigate = useNavigate()

  return (
    <main>
      <div
        className="mt-2 px-2 flex flex-col space-y-2 items-start "
      >
        <span className="text-3xl font-bold">Поступление на объект</span>
        <Button
          text="Добавить"
          onClick={() => navigate(INVOICE_OBJECT_MUTATION_ADD)}
        />
        {/* <Button text="Экспорт" onClick={() => {}}/> */}
      </div>
      <div className="grid grid-cols-1 gap-2 px-3 mt-2">
        {tableDataQuery.isSuccess && tableData.map((value, index) =>
          <div key={index} className="grid grid-cols-2 gap-2 px-3 py-2 bg-gray-800 text-white rounded-md overflow-auto">
            <div className="font-bold">Код поступления</div>
            <div>{value.deliveryCode}</div>
            <div className="font-bold">Супервайзер</div>
            <div>{value.supervisorName}</div>
            <div className="font-bold">Объект</div>
            <div>{value.objectName} ({objectTypeIntoRus(value.objectType)})</div>
            <div className="font-bold">Бригада</div>
            <div>{value.teamNumber}</div>
            <div className="font-bold">Дата поступления</div>
            <div>{value.dateOfInvoice.toString().substring(0, 10)}</div>
            <div className="col-span-2 flex space-x-2">
              <Button
                text="Подробнее"
                onClick={() => navigate(`/invoice/object/${value.id}`)}
              />
            </div>
          </div>
        )}
        {tableDataQuery.hasNextPage &&
          <div className="w-full py-4 flex justify-center">
            <div
              onClick={() => tableDataQuery.fetchNextPage()}
              className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
            >
              {tableDataQuery.isLoading && <LoadingDots height={30} />}
              {!tableDataQuery.isLoading && "Загрузить еще"}
            </div>
          </div>
        }

      </div>
    </main>
  )
}
