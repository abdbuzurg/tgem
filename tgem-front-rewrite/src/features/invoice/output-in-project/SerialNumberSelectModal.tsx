import Modal from "@shared/components/Modal";
import IReactSelectOptions from "@shared/types/ReactSelectOptions";
import Select from 'react-select'
import { useEffect, useState } from "react";
import Button from "@shared/ui/Button";
import { useQuery } from "@tanstack/react-query";
import { getSerialNumberCodesByMaterialID } from "./api";

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
  materialID: number
  addSerialNumbersToInvoice: (serailNumbes: string[]) => void
  alreadySelectedSerialNumers: string[]
}

export default function SerialNumberSelectModal({
  setShowModal,
  addSerialNumbersToInvoice,
  materialID,
  alreadySelectedSerialNumers,
}: Props) {

  //Available serial number Logic
  const [availableSerialNumbers, setAvailableSerialNumbers] = useState<IReactSelectOptions<string>[]>([])
  const [selectedAvailableSerialNumber, setSelectedAvailableSerialNumber] = useState<IReactSelectOptions<string>>({ label: "", value: "" })

  const codesQuery = useQuery<string[], Error, string[]>({
    queryKey: [`serial-number-for-${materialID}`],
    queryFn: () => getSerialNumberCodesByMaterialID(materialID),
  })
  useEffect(() => {
    if (codesQuery.isSuccess && codesQuery.data) {
      setAvailableSerialNumbers(
        codesQuery.data.
          filter((value) => alreadySelectedSerialNumers.indexOf(value) == -1).
          map<IReactSelectOptions<string>>((value) => ({ label: value, value: value }))
      )
    }
  }, [codesQuery.data, alreadySelectedSerialNumers, codesQuery.isSuccess])

  const onAddClick = () => {
    setPickedSerialNumbers([
      selectedAvailableSerialNumber,
      ...pickedSerialNumbers,
    ])

    setAvailableSerialNumbers(
      availableSerialNumbers.filter((value) => value.value != selectedAvailableSerialNumber.value)
    )

    setSelectedAvailableSerialNumber({ label: "", value: "" })
  }

  //Selected serial number logic (from available ones)
  const [pickedSerialNumbers, setPickedSerialNumbers] = useState<IReactSelectOptions<string>[]>(() => {
    if (!alreadySelectedSerialNumers) return []
    return alreadySelectedSerialNumers.map(value => ({ value: value, label: value }))
  })
  const [chosenSerialNumber, setChosenSerialNumber] = useState<IReactSelectOptions<string>>({ label: "", value: "" })

  const onDeleteClick = () => {
    setPickedSerialNumbers(
      pickedSerialNumbers.filter((value) => value.value != chosenSerialNumber.value)
    )

    setAvailableSerialNumbers([
      chosenSerialNumber,
      ...availableSerialNumbers,
    ])

    setChosenSerialNumber({ label: "", value: "" })
  }

  return (
    <Modal setShowModal={setShowModal}>
      <div className="flex flex-col space-y-3 py-2">
        <div>
          <span className="text-xl font-semibold">Выбор серийных номеров для отпуска</span>
        </div>
        <div>
          <label htmlFor="serialNumbers">Доступные серийные номера</label>
          <div className="flex space-x-2 items-center">
            <div className="w-full">
              <Select
                className="basic-single"
                classNamePrefix="select"
                isSearchable={true}
                isClearable={true}
                menuPosition="fixed"
                name={"materials"}
                placeholder={""}
                value={selectedAvailableSerialNumber}
                options={availableSerialNumbers}
                onChange={(value) => setSelectedAvailableSerialNumber({ label: value?.label ?? "", value: value?.value ?? "" })}
              />
            </div>
            <Button
              text="Добавить"
              disabled={selectedAvailableSerialNumber.label == "" && selectedAvailableSerialNumber.value == ""}
              onClick={onAddClick}
            />
          </div>
        </div>
        <div>
          <label htmlFor="serialNumbers">Выбранные серийные номера</label>
          <div className="flex space-x-2 items-center">
            <div className="w-full">
              <Select
                className="basic-single"
                classNamePrefix="select"
                isSearchable={true}
                isClearable={true}
                menuPosition="fixed"
                name={"materials"}
                placeholder={""}
                value={chosenSerialNumber}
                options={pickedSerialNumbers}
                onChange={(value) => setChosenSerialNumber({ label: value?.label ?? "", value: value?.value ?? "" })}
              />
            </div>
            <Button
              text="Удалить"
              buttonType="delete"
              disabled={chosenSerialNumber.label == "" && chosenSerialNumber.value == ""}
              onClick={onDeleteClick}
            />
          </div>
        </div>
        <div>
          <Button text="Привязать к материалу" buttonType="default" onClick={() => addSerialNumbersToInvoice(pickedSerialNumbers.map<string>(value => value.value))} />
        </div>
      </div>
    </Modal>
  )
}
