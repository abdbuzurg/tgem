import { useEffect, useState } from "react";
import Modal from "@shared/components/Modal";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { InvoiceCorrectionMaterial, InvoiceCorrectionMaterialMutation, InvoiceCorrectionOperation, InvoiceCorrectionPaginatedView, createInvoiceCorrection, getInvoiceMaterialsForCorrect, getOperationsForCorrect } from "./api";
import Button from "@shared/ui/Button";
import MaterialCorrectionModal from "./MaterialCorrectionModal";
import { IInvoiceObject } from "@entities/invoice/types/invoiceObject";
import LoadingDots from "@shared/ui/LoadingDots";
import OperationCorrectionModal from "./OperationCorrectionModal";
import toast from "react-hot-toast";

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
  invoiceObject: InvoiceCorrectionPaginatedView
}

export default function CorrectionModal({
  setShowModal,
  invoiceObject,
}: Props) {
  const [invoiceMaterialsForCorrection, setInvoiceMaterialsForCorrections] = useState<InvoiceCorrectionMaterial[]>([])
  const materialsForCorrectionQuery = useQuery<InvoiceCorrectionMaterial[], Error, InvoiceCorrectionMaterial[]>({
    queryKey: [`invoice-correction-materials`, invoiceObject.id],
    queryFn: () => getInvoiceMaterialsForCorrect(invoiceObject.id),
  })
  useEffect(() => {
    if (materialsForCorrectionQuery.isSuccess && materialsForCorrectionQuery.data) {
      setInvoiceMaterialsForCorrections(materialsForCorrectionQuery.data)
    }
  }, [materialsForCorrectionQuery.data, materialsForCorrectionQuery.isSuccess])


  const [showCorrectionMaterialModal, setShowCorrectionMaterialModal] = useState(false)
  const [selectedMaterial, setSelectedMaterial] = useState<InvoiceCorrectionMaterial>()
  const [selectedInvoiceCorrectionMaterialIndex, setSelectedInvoiceCorrectionMaterialIndex] = useState<number>(0)
  const onMaterialModalToggle = (materialData: InvoiceCorrectionMaterial, index: number) => {
    setSelectedInvoiceCorrectionMaterialIndex(index)
    setSelectedMaterial(materialData)
    setShowCorrectionMaterialModal(true)
  }

  const onMaterialDelete = (index: number) => {
    const data = invoiceMaterialsForCorrection.filter((_, idx) => index != idx)
    setInvoiceMaterialsForCorrections(data)
  }

  const materialCorrection = (index: number, correction: InvoiceCorrectionMaterial) => {
    if (index == -1) {
      if (invoiceMaterialsForCorrection.findIndex(val => val.materialID == correction.materialID) != -1) {
        toast.error("Такой материал уже в списке")
        return
      }

      invoiceMaterialsForCorrection.push(correction)
    } else {
      invoiceMaterialsForCorrection[index] = correction
    }

    setInvoiceMaterialsForCorrections(invoiceMaterialsForCorrection)
    setShowCorrectionMaterialModal(false)
  }

  const [operationsForCorrection, setOperationsForCorrection] = useState<InvoiceCorrectionOperation[]>([])
  const operationsForCorrectionQuery = useQuery<InvoiceCorrectionOperation[], Error, InvoiceCorrectionOperation[]>({
    queryKey: ["invoice-correction-operations", invoiceObject.id],
    queryFn: () => getOperationsForCorrect(invoiceObject.id)
  })
  useEffect(() => {
    if (operationsForCorrectionQuery.isSuccess && operationsForCorrectionQuery.data) {
      setOperationsForCorrection(operationsForCorrectionQuery.data)
    }
  }, [operationsForCorrectionQuery.data, operationsForCorrectionQuery.isSuccess])

  const [showCorrectionOperationModal, setShowCorrectionOperationModal] = useState(false)
  const [selectedOperation, setSelectedOperation] = useState<InvoiceCorrectionOperation>()
  const [selectedInvoiceCorrectionOperationIndex, setSelectedInvoiceCorrectionOperationIndex] = useState<number>(0)
  const onOperationModalToggle = (operationData: InvoiceCorrectionOperation, index: number) => {
    setSelectedOperation(operationData)
    setSelectedInvoiceCorrectionOperationIndex(index)
    setShowCorrectionOperationModal(true)
  }

  const onOperationDelete = (index: number) => {
    const data = operationsForCorrection.filter((_, idx) => index != idx)
    setOperationsForCorrection(data)
  }

  const operationCorrection = (index: number, correction: InvoiceCorrectionOperation) => {

    if (index == -1) {
      if (operationsForCorrection.findIndex(val => val.operationID == correction.operationID) != -1) {
        toast.error("Такая услуга уже в списке")
        return
      }

      operationsForCorrection.push(correction)
    } else {
      operationsForCorrection[index] = correction
    }

    setOperationsForCorrection(operationsForCorrection)
    setShowCorrectionOperationModal(false)
  }

  const queryClient = useQueryClient()
  const createInvoiceCorrectionMutation = useMutation<IInvoiceObject, Error, InvoiceCorrectionMaterialMutation>({
    mutationFn: createInvoiceCorrection
  })

  const onCorrectionSubmit = () => {

    createInvoiceCorrectionMutation.mutate({
      details: {
        id: invoiceObject.id,
        dateOfCorrection: new Date()
      },
      items: invoiceMaterialsForCorrection,
      operations: operationsForCorrection,
    }, {
      onSuccess: () => {
        setShowModal(false)
        queryClient.invalidateQueries(["invoice-correction"])
      }
    })
  }

  return (
    <Modal setShowModal={setShowModal} bigModal>
      <div className="flex flex-col space-y-2 px-2">
        <span className="text-2xl font-bold">Корректировка прихода {invoiceObject.deliveryCode}</span>
        <div className="flex space-x-4">
          <div className="flex flex-col space-y-1">
            <span className="font-bold">Супервайзер</span>
            <span>{invoiceObject.supervisorName}</span>
          </div>
          <div className="flex flex-col space-y-1">
            <span className="font-bold text-l">Объект</span>
            <span>{invoiceObject.objectName}</span>
          </div>
          <div className="flex flex-col space-y-1">
            <span className="font-bold text-l">Бригада</span>
            <span>{invoiceObject.teamLeaderName}</span>
          </div>
          <div className="flex flex-col space-y-1">
            <span className="font-bold text-l">Дата накладной</span>
            <span>{invoiceObject.dateOfInvoice.toString().substring(0, 10)}</span>
          </div>
          <div className="flex items-center">
            <div
              onClick={() => onCorrectionSubmit()}
              className="text-center text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
            >
              {createInvoiceCorrectionMutation.isLoading ? <LoadingDots height={30} /> : "Опубликовать"}
            </div>
          </div>
        </div>
        <div className="flex flex-col mt-3">
          <span className="text-xl font-bold mb-1">Материалы для корректировки</span>
          <div className="grid grid-cols-4 gap-1 border-b-2 border-b-black font-bold text-l">
            <div className="px-2 py-1 flex items-center">Наименование</div>
            <div className="px-2 py-1 flex items-center">Кол-во</div>
            <div className="px-2 py-1 flex items-center">Примичание</div>
            <div className="px-2 py-1 font-normal">
              <Button
                onClick={() => onMaterialModalToggle({
                  invoiceMaterialID: 0,
                  materialName: "",
                  materialID: 0,
                  materialAmount: 0,
                  materialAvailableAmount: 0,
                  materialUnit: "",
                  notes: "",
                }, -1)}
                text="Добавить"
              />
            </div>
          </div>
          {materialsForCorrectionQuery.isLoading &&
            <div className="grid grid-cols-4 gap-1 border-b-2 border-b-black font-bold text-l">
              <div className="px-2 py-1 col-span-4"><LoadingDots /></div>
            </div>
          }
          {invoiceMaterialsForCorrection.map((row, index) =>
            <div className="grid grid-cols-4 gap-1 border-b border-b-black" key={index}>
              <div className="px-2 py-1 flex items-center">{row.materialName}</div>
              <div className="px-2 py-1 flex items-center">{row.materialAmount}</div>
              <div className="px-2 py-1 flex items-center">{row.notes}</div>
              <div className="px-2 py-1 flex space-x-2">
                <Button onClick={() => onMaterialModalToggle(row, index)} text="Изменить" />
                <Button onClick={() => onMaterialDelete(index)} text="Удалить" buttonType="delete" />
              </div>
            </div>
          )}
        </div>
        <div className="flex flex-col mt-3">
          <span className="text-xl font-bold mb-1">Услуги для корректировки</span>
          <div className="grid grid-cols-4 gap-1 border-b-2 border-b-black font-bold text-l">
            <div className="px-2 py-1 flex items-center">Наименование услуги</div>
            <div className="px-2 py-1 flex items-center">Кол-во</div>
            <div className="px-2 py-1 flex items-center">Привязанный материал</div>
            <div className="px-2 py-1 font-normal">
              <Button
                onClick={() => onOperationModalToggle({
                  materialName: "",
                  operationID: 0,
                  operationName: "",
                  amount: 0,
                }, -1)}
                text="Добавить"
              />
            </div>
          </div>
          {operationsForCorrectionQuery.isLoading &&
            <div className="grid grid-cols-4 gap-1 border-b-2 border-b-black font-bold text-l">
              <div className="px-2 py-1 col-span-4"><LoadingDots /></div>
            </div>
          }
          {operationsForCorrection.map((row, index) =>
            <div className="grid grid-cols-4 gap-1 border-b border-b-black" key={index}>
              <div className="px-2 py-1 flex items-center">{row.operationName}</div>
              <div className="px-2 py-1 flex items-center">{row.amount}</div>
              <div className="px-2 py-1 flex items-center">{row.materialName}</div>
              <div className="px-2 py-1 flex space-x-2">
                <Button onClick={() => onOperationModalToggle(row, index)} text="Изменить" />
                <Button onClick={() => onOperationDelete(index)} text="Удалить" buttonType="delete" />
              </div>
            </div>
          )}
        </div>
      </div>
      {showCorrectionMaterialModal &&
        <MaterialCorrectionModal
          teamID={invoiceObject.teamID}
          correctionIndex={selectedInvoiceCorrectionMaterialIndex}
          setShowModal={setShowCorrectionMaterialModal}
          materialData={selectedMaterial!}
          correctionFunction={materialCorrection}
        />
      }
      {showCorrectionOperationModal &&
        <OperationCorrectionModal
          setShowModal={setShowCorrectionOperationModal}
          operationData={selectedOperation!}
          correctionFunction={operationCorrection}
          correctionIndex={selectedInvoiceCorrectionOperationIndex}
          teamID={invoiceObject.teamID}
        />
      }
    </Modal>
  )
}
