import { useQuery } from "@tanstack/react-query"
import { InvoiceObjectDescriptiveData, InvoiceObjectView, getInvoiceObjectDescriptiveDataByID } from "@features/invoice/object/api"
import { objectTypeIntoRus } from "@shared/lib/data/objectStatuses"
import Modal from "@shared/components/Modal"
import LoadingDots from "@shared/ui/LoadingDots"
import React from "react"

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
  data: InvoiceObjectView
}

export default function ShowInvoiceObjectDetails({ setShowModal, data }: Props) {

  const fullDataQuery = useQuery<InvoiceObjectDescriptiveData, Error, InvoiceObjectDescriptiveData>({
    queryKey: [`invoice-object-${data.id}`],
    queryFn: () => getInvoiceObjectDescriptiveDataByID(data.id),
  })

  return (
    <Modal setShowModal={setShowModal} bigModal>
      <div className="mb-2 px-1">
        <h3 className="text-2xl font-medium text-gray-800">Детали расходы {data.deliveryCode}</h3>
      </div>
      <div className="flex w-full space-x-8 max-h-[70vh] px-2">
        <div className="flex flex-col w-[25%]">
          <div className="flex flex-col space-y-3">
            <div className="flex flex-col">
              <p className="text-sm font-bold">Объект</p>
              <p className="text-xl italic">{`${data.objectName}(${objectTypeIntoRus(data.objectType)})`}</p>
            </div>
            <div className="flex flex-col">
              <p className="text-sm font-bold">Бригада</p>
              <p className="text-xl italic">{data.teamNumber}</p>
            </div>
            <div className="flex flex-col">
              <p className="text-sm font-bold">Дата</p>
              <p className="text-xl italic">{data.dateOfInvoice.toString().substring(0, 10)}</p>
            </div>
          </div>
        </div>
        <div className="w-full px-2">
          <div className="py-1">
            <div className="py-1">
              <span className="font-bold text-xl">Материалы расхода</span>
            </div>
            {fullDataQuery.isLoading &&
              <>
                <LoadingDots />
              </>
            }
            {fullDataQuery.isSuccess && 
              <>
                <div className="grid grid-cols-5 gap-2 font-semibold border-b-2">
                  <div className="col-span-2">Наименование</div>
                  <div>Ед. Изм.</div>
                  <div>Количество</div>
                  <div>Примичание</div>
                </div>
                <div className="grid grid-cols-5 gap-2 max-h-[30vh] overflow-y-auto py-2">
                  {fullDataQuery.data.materialsWithoutSN.map((value, index) =>
                    <React.Fragment key={index}>
                      <div className="col-span-2">{value.materialName}</div>
                      <div>{value.materialUnit}</div>
                      <div>{value.amount}</div>
                      <div>{value.notes}</div>
                    </React.Fragment>
                  )}
                </div>
              </>
            }
          </div>
        </div>
      </div>
    </Modal>
  )
}
