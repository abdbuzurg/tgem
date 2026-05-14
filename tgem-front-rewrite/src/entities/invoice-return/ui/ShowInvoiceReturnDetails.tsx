import { useQuery } from "@tanstack/react-query"
import { InvoiceMaterialViewWithSerialNumbers, InvoiceMaterialViewWithoutSerialNumbers } from "@entities/invoice/types/invoiceMaterial"
import { IInvoiceReturnView } from "../types"
import { Fragment, useState } from "react"
import Modal from "@shared/components/Modal"
import IconButton from "@shared/components/IconButtons"
import { FaBarcode } from "react-icons/fa"
import LoadingDots from "@shared/ui/LoadingDots"
import SerialNumbersShowModal from "@entities/invoice/ui/SerialNumberShowModal"
import { getInvoiceReturnMaterilsWithSerialNumbersByID, getInvoiceReturnMaterilsWithoutSerialNumbersByID } from "../api"

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
  data: IInvoiceReturnView
}

export default function ShowInvoiceReturnDetails({ setShowModal, data }: Props) {

  const invoiceOutputMaterialsWithourSerialNumbersQuery = useQuery<InvoiceMaterialViewWithoutSerialNumbers[], Error>({
    queryKey: ["invoice-return-materials-without-serial-numbers", data.id],
    queryFn: () => getInvoiceReturnMaterilsWithoutSerialNumbersByID(data.id)
  })

  const invoiceOutputMaterialsWithSerialNumbersQuery = useQuery<InvoiceMaterialViewWithSerialNumbers[], Error>({
    queryKey: ["invoice-return-materials-with-serial-numbers", data.id],
    queryFn: () => getInvoiceReturnMaterilsWithSerialNumbersByID(data.id)
  })

  const [showSerialNumbers, setShowSerialNumbers] = useState(false)
  const [currentMaterial, setCurrentMaterial] = useState<InvoiceMaterialViewWithSerialNumbers>({
    id: 0,
    materialName: "",
    materialUnit: "",
    costM19: 0,
    amount: 0,
    notes: "",
    serialNumbers: [],
    isDefected: false,
  })

  return (
    <Modal setShowModal={setShowModal} bigModal>
      <div className="mb-2 px-1">
        <h3 className="text-2xl font-medium text-gray-800">Детали накладной {data.deliveryCode}</h3>
      </div>
      <div className="flex w-full space-x-8 max-h-[70vh] px-2">
        <div className="flex flex-col w-[25%]">
          <div className="flex flex-col space-y-3">
            {data.teamNumber && data.teamLeaderName &&
              <>
                <div className="flex flex-col">
                  <p className="text-sm font-bold">Номер Бригады</p>
                  <p className="text-xl italic">{data.teamNumber}</p>
                </div>
                <div className="flex flex-col">
                  <p className="text-sm font-bold">Бригадиры</p>
                  <p className="text-xl italic">{data.dateOfInvoice.toString().substring(0, 10)}</p>
                </div>
              </>
            }
            {data.objectName && data.objectSupervisorNames &&
              <>
                <div className="flex flex-col">
                  <p className="text-sm font-bold">Объект</p>
                  <p className="text-xl italic">{data.objectName}</p>
                </div>
                <div className="flex flex-col">
                  <p className="text-sm font-bold">Супервайзеры</p>
                  <p className="text-xl italic">{data.objectSupervisorNames.toString().substring(0, 10)}</p>
                </div>
              </>
            }
            <div className="flex flex-col">
              <p className="text-sm font-bold">Дата</p>
              <p className="text-xl italic">{data.dateOfInvoice.toString().substring(0, 10)}</p>
            </div>
          </div>
        </div>
        <div className="w-full px-2">
          <div className="py-1">
            <div className="py-1">
              <span className="font-bold text-xl">Материалы накладной с серийными номерами</span>
            </div>
            {invoiceOutputMaterialsWithSerialNumbersQuery.isLoading && <div className="col-span-6"><LoadingDots /></div>}
            {invoiceOutputMaterialsWithSerialNumbersQuery.isSuccess &&
              <>
                {invoiceOutputMaterialsWithSerialNumbersQuery.data.length == 0 &&
                  <div className="italic">Отсутствует</div>
                }
                {invoiceOutputMaterialsWithSerialNumbersQuery.data.length != 0 &&
                  <>
                    <div className="grid grid-cols-8 gap-2 font-semibold border-b-2">
                      <div className="col-span-2">Наименование</div>
                      <div>Ед. Изм.</div>
                      <div>Количество</div>
                      <div>Цена</div>
                      <div>Брак</div>
                      <div>Серийный #</div>
                      <div>Примичание</div>
                    </div>
                    <div className="grid grid-cols-8 gap-2 max-h-[30vh] overflow-y-auto py-2">
                      {invoiceOutputMaterialsWithSerialNumbersQuery.data.map((value, index) =>
                        <Fragment key={index}>
                          <div className="col-span-2">{value.materialName}</div>
                          <div>{value.materialUnit}</div>
                          <div>{value.amount}</div>
                          <div>{value.costM19}</div>
                          <div>{value.isDefected ? "Да" : "Нет" }</div>
                          <div>
                            <IconButton
                              icon={<FaBarcode
                                size="25px"
                                title={`Показать серийные номера`} />}
                              onClick={() => {
                                setCurrentMaterial(value)
                                setShowSerialNumbers(true)
                              }}
                            />
                          </div>
                          <div>{value.notes}</div>
                        </Fragment>
                      )}
                    </div>
                  </>
                }
              </>
            }
            {showSerialNumbers && <SerialNumbersShowModal setShowModal={setShowSerialNumbers} data={currentMaterial} />}
          </div>
          <div className="py-1">
            <div className="py-1">
              <span className="font-bold text-xl">Материалы накладной без серийного номера</span>
            </div>
            {invoiceOutputMaterialsWithourSerialNumbersQuery.isLoading && <div className="col-span-6"><LoadingDots /></div>}
            {invoiceOutputMaterialsWithourSerialNumbersQuery.isSuccess &&
              <>
                {invoiceOutputMaterialsWithourSerialNumbersQuery.data.length == 0 &&
                  <div className="italic">Отсутствует</div>
                }
                {invoiceOutputMaterialsWithourSerialNumbersQuery.data.length != 0 &&
                  <>
                    <div className="grid grid-cols-7 gap-2 font-semibold border-b-2">
                      <div className="col-span-2">Наименование</div>
                      <div>Ед. Изм.</div>
                      <div>Количество</div>
                      <div>Цена</div>
                      <div>Брак</div>
                      <div>Примичание</div>
                    </div>
                    <div className="grid grid-cols-7 gap-2 max-h-[30vh] overflow-y-auto py-2">
                      {invoiceOutputMaterialsWithourSerialNumbersQuery.data.map((value, index) =>
                        <Fragment key={index}>
                          <div className="col-span-2">{value.materialName}</div>
                          <div>{value.materialUnit}</div>
                          <div>{value.amount}</div>
                          <div>{value.costM19}</div>
                          <div>{value.isDefected ? "Да" : "Нет" }</div>
                          <div>{value.notes}</div>
                        </Fragment>
                      )}
                    </div>
                  </>
                }
              </>
            }
          </div>
        </div>
      </div>
    </Modal >
  )

}
