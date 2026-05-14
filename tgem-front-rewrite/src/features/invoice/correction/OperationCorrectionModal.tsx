import { useQuery } from "@tanstack/react-query";
import IReactSelectOptions from "@shared/types/ReactSelectOptions";
import Modal from "@shared/components/Modal";
import Select from 'react-select'
import { useEffect, useState } from "react";
import { InvoiceCorrectionOperation } from "./api";
import Input from "@shared/ui/Input";
import toast from "react-hot-toast";
import { InvoiceObjectOperations, getOperationsBasedOnTeamID } from "@features/invoice/object/api";

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
  operationData: InvoiceCorrectionOperation
  correctionIndex: number
  teamID: number
  correctionFunction: (index: number, correction: InvoiceCorrectionOperation) => void
}

export default function OperationCorrectionModal({
  setShowModal,
  operationData,
  correctionIndex,
  correctionFunction,
  teamID
}: Props) {

  const [correction, setCorrection] = useState<InvoiceCorrectionOperation>(operationData)

  const [selectedOperation, setSelectedOperation] = useState<IReactSelectOptions<number>>({
    label: "",
    value: 0,
  })
  useEffect(() => {
    if (allOperationsQuery.isSuccess && allOperationsQuery.data && selectedOperation.value != 0) {
      const operation = allOperationsQuery.data.find(val => val.operationID == selectedOperation.value)!
      setCorrection({
        ...correction,
        operationName: operation.operationName,
        operationID:  operation.operationID,
        materialName: operation.materialName,
      })
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedOperation])
  const [allOperations, setAllOperations] = useState<IReactSelectOptions<number>[]>([])
  const allOperationsQuery = useQuery<InvoiceObjectOperations[], Error, InvoiceObjectOperations[]>({
    queryKey: ["operations-in-team", teamID],
    queryFn: () => getOperationsBasedOnTeamID(teamID),
  })
  useEffect(() => {
    if (allOperationsQuery.isSuccess && allOperationsQuery.data) {
      setAllOperations(allOperationsQuery.data.map<IReactSelectOptions<number>>(val => ({
        label: val.operationName,
        value: val.operationID,
      })))

      if (operationData.operationID != 0) {
        const operation = allOperationsQuery.data.find(val => val.operationID == operationData.operationID)!
        setSelectedOperation({
          label: operation.operationName,
          value: operation.operationID,
        })
      }
    }
  }, [allOperationsQuery.data, allOperationsQuery.isSuccess, operationData.operationID])

  const submitChanges = () => {
    if (correction.operationID == 0) {
      toast.error("Не выбрана услуга")
      return
    }  

    if (correction.amount <= 0) {
      toast.error("Неправильно указано количество для услуги")
      return
    }
  
    correctionFunction(correctionIndex, correction)
  }

  return (
    <Modal setShowModal={setShowModal}>
      <div className="flex flex-col space-y-2">
        <span className="font-bold text-2xl">Корректировка услуги</span>
        <div className="flex flex-col space-y-1">
          <label className="font-bold" htmlFor="operations">Наименование</label>
          <Select
            className="basic-single"
            classNamePrefix="select"
            isSearchable={true}
            isClearable={true}
            name={"operations"}
            placeholder={""}
            value={selectedOperation}
            options={allOperations}
            onChange={(value) => {
              setSelectedOperation(value ?? {value: 0, label:  ""})
            }}
          />
          {selectedOperation.value != 0 && <span className="text-sm italic">Привязанный материал - {correction.materialName == "" ? "Отсутсвует" : correction.materialName}</span>}
        </div>
        <div className="flex space-x-2 items-center">
          <div className="flex flex-col space-y-1">
            <label className="font-bold">Количество</label>
            <Input
              name="amount"
              type="number"
              value={correction.amount}
              onChange={(e) => setCorrection({
                ...correction,
                amount: e.target.valueAsNumber
              })}
            />
          </div>
        </div>
        <div className="flex">
          <div
            onClick={() => submitChanges()}
            className="text-white py-2.5 px-5 rounded-lg bg-gray-800 hover:bg-gray-700 hover:cursor-pointer"
          >
            Опубликовать
          </div>
        </div>
      </div>
    </Modal>
  )
}
