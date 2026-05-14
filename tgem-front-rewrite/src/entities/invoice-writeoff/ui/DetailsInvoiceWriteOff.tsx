import { useQuery } from "@tanstack/react-query"
import { InvoiceMaterialViewWithoutSerialNumbers } from "@entities/invoice/types/invoiceMaterial"
import { IInvoiceWriteOffView } from "../types"
import Modal from "@shared/components/Modal"
import LoadingDots from "@shared/ui/LoadingDots"
import { getInvoiceWriteOffMaterilsWithoutSerialNumbersByID } from "../api"
import { Fragment } from "react"

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
  data: IInvoiceWriteOffView
}

export default function DetailsInvoiceWriteOff({
  setShowModal,
  data,
}: Props) {

  const invoiceInputMaterialsWithourSerialNumbersQuery = useQuery<InvoiceMaterialViewWithoutSerialNumbers[], Error>({
    queryKey: ["invoice-writeoff-materials-without-serial-numbers", data.id],
    queryFn: () => getInvoiceWriteOffMaterilsWithoutSerialNumbersByID(data.id)
  })

  return (
    <Modal setShowModal={setShowModal} bigModal>
      <div className="mb-2 px-1">
        <h3 className="text-2xl font-medium text-gray-800">Детали накладной {data.deliveryCode}</h3>
      </div>
      <div className="flex w-full space-x-8 max-h-[70vh] px-2">
        <div className="flex flex-col w-[25%]">
          <div className="flex flex-col space-y-3">
            <div className="flex flex-col">
              <p className="text-sm font-bold">Составитель</p>
              <p className="text-xl italic">{data.releasedWorkerName}</p>
            </div>
            {data.writeOffType == "loss-team" &&
              <div className="flex flex-col">
                <p className="text-sm font-bold">Списано из</p>
                <p className="text-xl italic">{data.writeOffLocationName}</p>
              </div>
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
              <span className="font-bold text-xl">Материалы накладной без серийного номера</span>
            </div>
            {invoiceInputMaterialsWithourSerialNumbersQuery.isLoading && <div className="col-span-6"><LoadingDots /></div>}
            {invoiceInputMaterialsWithourSerialNumbersQuery.isSuccess &&
              <>
                {invoiceInputMaterialsWithourSerialNumbersQuery.data.length == 0 &&
                  <div className="italic">Отсутствует</div>
                }
                {invoiceInputMaterialsWithourSerialNumbersQuery.data.length != 0 &&
                  <>
                    <div className="grid grid-cols-6 gap-2 font-semibold border-b-2">
                      <div className="col-span-2">Наименование</div>
                      <div>Ед. Изм.</div>
                      <div>Количество</div>
                      <div>Цена</div>
                      <div>Примичание</div>
                    </div>
                    <div className="grid grid-cols-6 gap-2 max-h-[30vh] overflow-y-auto py-2">
                      {invoiceInputMaterialsWithourSerialNumbersQuery.data.map((value, index) =>
                        <Fragment key={index}>
                          <div className="col-span-2">{value.materialName}</div>
                          <div>{value.materialUnit}</div>
                          <div>{value.amount}</div>
                          <div>{value.costM19}</div>
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
    </Modal>
  )
}
