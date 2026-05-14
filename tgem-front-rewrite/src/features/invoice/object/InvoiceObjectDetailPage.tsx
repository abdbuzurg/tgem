import { useParams } from "react-router-dom"
import { InvoiceObjectDescriptiveData, getInvoiceObjectDescriptiveDataByID } from "@features/invoice/object/api"
import { useQuery } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import LoadingDots from "@shared/ui/LoadingDots"
import IconButton from "@shared/components/IconButtons"
import { FaBarcode } from "react-icons/fa"
import SerialNumbersShowModal from "@entities/invoice/ui/SerialNumberShowModal"
import { InvoiceMaterialViewWithSerialNumbers } from "@entities/invoice/types/invoiceMaterial"
import { objectTypeIntoRus } from "@shared/lib/data/objectStatuses"

export default function InvoiceObjectDetails() {

  const { id } = useParams()
  const [displayData, setDisplayData] = useState<InvoiceObjectDescriptiveData>({
    invoiceData: {
      id: 0,
      deliveryCode: "",
      supervisorName: "",
      objectType: "",
      objectName: "",
      teamNumber: "",
      dateOfInvoice: new Date(),
      confirmedByOperator: false,
    },
    materialsWithSN: [],
    materialsWithoutSN: [],
  })

  const fullDataQuery = useQuery<InvoiceObjectDescriptiveData, Error, InvoiceObjectDescriptiveData>({
    queryKey: [`invoice-object-${id}`],
    queryFn: () => getInvoiceObjectDescriptiveDataByID(parseInt(id ?? "")),
    enabled: id != "",
  })

  useEffect(() => {
    if (fullDataQuery.isSuccess && fullDataQuery.data) {
      setDisplayData(fullDataQuery.data)
    }
  }, [fullDataQuery.data, fullDataQuery.isSuccess])

  const [current, setCurrent] = useState<InvoiceMaterialViewWithSerialNumbers>({
    id: 0,
    materialName: "",
    materialUnit: "",
    isDefected: false,
    costM19: 0,
    amount: 0,
    notes: "",
    serialNumbers: [],
  })
  const [showSerialNumberModal, setShowSerialNumberModal] = useState(false)

  return (
    <main>
      <div className="my-1 px-2 flex flex-col space-y-2 items-start ">
        <span className="text-2xl font-bold">Детальная информация посутпления</span>
      </div>
      {fullDataQuery.isLoading && <LoadingDots />}
      {fullDataQuery.isSuccess &&
        <>
          <div className="mt-1 px-3">
            <div className="bg-gray-800 text-white rounded-md px-3 py-2">
              <span className="text-xl">Основная информация поступления</span>
              <div className="my-2 grid grid-cols-2 gap-2">
                <div className="font-bold">Код поступления</div>
                <div>{displayData.invoiceData.deliveryCode}</div>
                <div className="font-bold">Супервайзер</div>
                <div>{displayData.invoiceData.supervisorName}</div>
                <div className="font-bold">Объект</div>
                <div>{displayData.invoiceData.objectName} ({objectTypeIntoRus(displayData.invoiceData.objectType)})</div>
                <div className="font-bold">Бригада</div>
                <div>{displayData.invoiceData.teamNumber}</div>
                <div className="font-bold">Дата поступления</div>
                <div>{displayData.invoiceData.dateOfInvoice.toString().substring(0, 10)}</div>
                <div className="font-bold">Подтверждено оператором</div>
                <div>{displayData.invoiceData.confirmedByOperator ? "ДА" : "НЕТ"}</div>
              </div>
            </div>
          </div>
          <div className="mt-1 px-3 flex flex-col space-y-2">
            {displayData.materialsWithoutSN.map((val) =>
              <div className="bg-gray-800 text-white rounded-md px-3 py-2">
                <span className="text-xl">Поступившие метиралы без серийного номера</span>
                <div className="flex flex-col space-y-2">
                  <div className="my-2 grid grid-cols-2 border border-white px-2 py-1 gap-1" key={val.id}>
                    <div className="font-bold">Наименование</div>
                    <div>{val.materialName}</div>
                    <div className="font-bold">Количество</div>
                    <div>{val.amount}</div>
                    <div className="font-bold">Примечание</div>
                    <div>{val.notes}</div>
                  </div>
                </div>
              </div>
            )}
            {displayData.materialsWithSN.map((val) =>
              <div className="bg-gray-800 text-white rounded-md px-3 py-2">
                <span className="text-xl">Поступившие метиралы c серийного номера</span>
                <div className="flex flex-col space-y-2">
                  <div className="my-2 grid grid-cols-2 border border-white px-2 py-1 gap-1" key={val.id}>
                    <div className="font-bold">Наименование</div>
                    <div>{val.materialName}</div>
                    <div className="font-bold">Количество</div>
                    <div>{val.amount}</div>
                    <div className="font-bold">Серийные номера</div>
                    <div>
                      <IconButton
                        icon={<FaBarcode
                          size="25px"
                          title={`Показать серийные номера`} />}
                        onClick={() => {
                          setCurrent(val)
                          setShowSerialNumberModal(true)
                        }}
                      />
                    </div>
                    <div className="font-bold">Примечание</div>
                    <div>{val.notes}</div>
                  </div>
                </div>
              </div>
            )}
          </div>
        </>
      }
      {showSerialNumberModal && <SerialNumbersShowModal setShowModal={setShowSerialNumberModal} data={current} />}

    </main>
  )
}
