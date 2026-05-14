import { useEffect, useState } from "react";
import IReactSelectOptions from "@shared/types/ReactSelectOptions";
import Modal from "@shared/components/Modal";
import Select from 'react-select'
import { useQuery } from "@tanstack/react-query";
import Input from "@shared/ui/Input";
import { InvoiceCorrectionMaterial } from "./api";
import toast from "react-hot-toast";
import { InvoiceObjectTeamMaterialData, getMaterialsDataFromTeam,  } from "@features/invoice/object/api"

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
  correctionIndex: number
  teamID: number
  materialData: InvoiceCorrectionMaterial
  correctionFunction: (index: number, correction: InvoiceCorrectionMaterial) => void
}

export default function MaterialCorrectionModal({
  setShowModal,
  correctionIndex,
  correctionFunction,
  materialData,
  teamID,
}: Props) {

  const [selectedMaterial, setSelectedMaterial] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [availableMaterials, setAvailableMaterials] = useState<IReactSelectOptions<number>[]>([])

  const materialQuery = useQuery<InvoiceObjectTeamMaterialData[], Error, InvoiceObjectTeamMaterialData[]>({
    queryKey: ["materials-in-team", teamID],
    queryFn: () => getMaterialsDataFromTeam(teamID),
  })

  useEffect(() => {
    if (materialQuery.isSuccess && materialQuery.data) {

      setAvailableMaterials([
        ...materialQuery.data.map<IReactSelectOptions<number>>(val => ({
          value: val.materialID,
          label: val.materialName,
        }))
      ])

      if (materialData.materialID != 0) {
        const material = materialQuery.data.find((val) => val.materialID == materialData.materialID)!
        setSelectedMaterial({
          value: material.materialID,
          label: material.materialName,
        })
      }
    }
  }, [materialQuery.data, materialData.materialID, materialQuery.isSuccess])

  useEffect(() => {
    if (selectedMaterial.value == 0) {
      return
    }

    if (materialQuery.isSuccess && materialQuery.data) {
      const material = materialQuery.data.find((val) => val.materialID == selectedMaterial.value)!
      setCorrection({
        ...correction,
        materialID: material.materialID,
        materialUnit: material.materialUnit,
        materialName: material.materialName,
        materialAvailableAmount: material.amount,
      })
    }

  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedMaterial, materialQuery.data, materialQuery.isSuccess])

  const [correction, setCorrection] = useState<InvoiceCorrectionMaterial>(materialData)

  const resetCorrection = () => setCorrection({
    invoiceMaterialID: 0,
    materialAmount: 0,
    materialAvailableAmount: 0,
    materialID: 0,
    materialName: "",
    notes: "",
    materialUnit: "",
  })

  // const [availableSerialNumbers, setAvaiableSerialNumbers] = useState<IReactSelectOptions<string>[]>([])
  // const [selectedSerialNumber, setSelectedSerialNumber] = useState<IReactSelectOptions<string>>({ label: "", value: "" })
  // const [alreadySelectedSerialNumbers, setAlreadySelectedSerialNumbers] = useState<IReactSelectOptions<string>[]>([])
  // const [toBeDeletedSerialNumber, setToBeDeletedSerialNumber] = useState<IReactSelectOptions<string>>({ label: "", value: "" })

  // const availableSerialNumbersQuery = useQuery<string[], Error, string[]>({
  //   queryKey: [`serial-numbers-of-material-${selectedMaterial.value}`],
  //   queryFn: () => getSerialNumbersOfMaterialInTeam(selectedMaterial.value, teamID),
  //   enabled: correction.hasSerialNumbers,
  // })
  // useEffect(() => {

  //   if (availableSerialNumbersQuery.isSuccess && availableSerialNumbersQuery.data) {
  //     setAvaiableSerialNumbers([
  //       ...availableSerialNumbersQuery.data.map<IReactSelectOptions<string>>((val) => ({ value: val, label: val }))
  //     ])
  //   }

  // }, [availableSerialNumbersQuery.data])

  // const addToSerialNumberList = () => {
  //   const index = alreadySelectedSerialNumbers.findIndex((val) => selectedSerialNumber.value == val.value)
  //   if (index != -1) {
  //     toast.error("Данный серийний код уже в списке")
  //     return
  //   }
  //   setAlreadySelectedSerialNumbers([selectedSerialNumber, ...alreadySelectedSerialNumbers])
  //   setSelectedSerialNumber({ label: "", value: "" })
  // }

  // const deleteSerialNumberFromList = () => {
  //   setAlreadySelectedSerialNumbers([...alreadySelectedSerialNumbers.filter((val) => val.value != toBeDeletedSerialNumber.value)])
  //   setToBeDeletedSerialNumber({ label: "", value: "" })
  // }

  const submitChanges = () => {
    if (selectedMaterial.value == 0) {
      toast.error("Не выбран материал")
      return
    }

    if (correction.materialAmount > correction.materialAvailableAmount) {
      toast.error(`Кол-во материала превышает доступное количество: ${correction.materialAmount} > ${correction.materialAvailableAmount}`)
      return
    }

    correctionFunction(correctionIndex, correction)
  }

  return (
    <Modal setShowModal={setShowModal}>
      <div className="flex flex-col space-y-2">
        <span className="font-bold text-2xl">Корректировка материла</span>
        <div className="flex flex-col space-y-1">
          <label className="font-bold">Наименование</label>
          <Select
            className="basic-single"
            classNamePrefix="select"
            isSearchable={true}
            isClearable={true}
            name={"materials"}
            placeholder={""}
            value={selectedMaterial}
            options={availableMaterials}
            onChange={(value) => {
              setSelectedMaterial({
                value: value?.value ?? 0,
                label: value?.label ?? "",
              })
              resetCorrection()
            }}
          />
        </div>
        <div className="flex space-x-2 items-center">
          <div className="flex flex-col space-y-1">
            <label className="font-bold">Количество</label>
            <Input
              name="amount"
              type="number"
              value={correction.materialAmount}
              onChange={(e) => setCorrection({
                ...correction,
                materialAmount: e.target.valueAsNumber
              })}
            />
          </div>
          <span>Доступно - {correction.materialAvailableAmount} {correction.materialUnit}</span>
        </div>
        {/* {correction.hasSerialNumbers && */}
        {/*   <div className="flex flex-col space-y-4"> */}
        {/*     <div className="flex space-x-2"> */}
        {/*       <div className="flex flex-col space-y-1"> */}
        {/*         <label>Доступные серийные номера</label> */}
        {/*         <Select */}
        {/*           className="basic-single text-black" */}
        {/*           classNamePrefix="select" */}
        {/*           isSearchable={true} */}
        {/*           isClearable={true} */}
        {/*           name={"material-cost-material-select"} */}
        {/*           placeholder={""} */}
        {/*           value={selectedSerialNumber} */}
        {/*           options={availableSerialNumbers} */}
        {/*           onChange={(value) => setSelectedSerialNumber({ */}
        {/*             label: value?.label ?? "", */}
        {/*             value: value?.value ?? "", */}
        {/*           })} */}
        {/*         /> */}
        {/*       </div> */}
        {/*       <Button onClick={() => addToSerialNumberList()} text="Выбрать" /> */}
        {/*     </div> */}
        {/*     <div className="flex space-x-2"> */}
        {/*       <div className="flex flex-col space-y-1"> */}
        {/*         <label>Выбранные серийные материалы</label> */}
        {/*         <Select */}
        {/*           className="basic-single text-black" */}
        {/*           classNamePrefix="select" */}
        {/*           isSearchable={true} */}
        {/*           isClearable={true} */}
        {/*           name={"material-cost-material-select"} */}
        {/*           placeholder={""} */}
        {/*           value={toBeDeletedSerialNumber} */}
        {/*           options={alreadySelectedSerialNumbers} */}
        {/*           onChange={(value) => setToBeDeletedSerialNumber({ */}
        {/*             label: value?.label ?? "", */}
        {/*             value: value?.value ?? "", */}
        {/*           })} */}
        {/*         /> */}
        {/*       </div> */}
        {/*       <Button onClick={() => deleteSerialNumberFromList()} text="Удалить" buttonType="delete" /> */}
        {/*     </div> */}
        {/*   </div> */}
        {/* } */}
        <div>
          <label className="font-bold">Примичание</label>
          <textarea
            className="w-full"
            value={correction.notes}
            onChange={(e) => setCorrection({ ...correction, notes: e.target.value })}
          />
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
