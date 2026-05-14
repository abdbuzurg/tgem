import { useState } from "react";
import Modal from "@shared/components/Modal";
import Input from "@shared/ui/Input";
import Button from "@shared/ui/Button";
import IReactSelectOptions from "@shared/types/ReactSelectOptions";
import Select from 'react-select'
import { toast } from "react-hot-toast";


interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
  addSerialNumbersToInvoice: (serialNumbers: string[]) => void
  availableSerialNumber: string[]
}

export default function SerialNumberAddModal({ setShowModal, addSerialNumbersToInvoice, availableSerialNumber }:Props) {
  const [serialNumber, setSerialNumber] = useState("")
  const [serialNumbers, setSerialNumbers] = useState<IReactSelectOptions<string>[]>(() => {
    if (!availableSerialNumber) return []
    return [
      ...availableSerialNumber.map<IReactSelectOptions<string>>(value => ({label: value, value: value}))
    ]
  })
  const [selectedSerialNumber, setSelectedSerialNumber] = useState<IReactSelectOptions<string>>({label:"", value: ""})

  const onAddClick = () => {
  
    if (serialNumber == "") {
      toast.error("Укажите серийный номер")
      return
    }

    const exists = serialNumbers.some((value) => serialNumber == value.value)
    if (exists) {
      toast.error("Такой серийный номер уже добавлен в список")
      return
    }
    setSerialNumbers([
      {
        label: serialNumber, 
        value: serialNumber
      },
      ...serialNumbers, 
    ])
    setSerialNumber("")
  }

  const onDeleteClick = () => {
    setSerialNumbers(serialNumbers.filter((serialNumber) => serialNumber.value != selectedSerialNumber.value))
    setSelectedSerialNumber({label: "", value: ""})
  }

  return (
    <Modal setShowModal={setShowModal}>
      <div className="flex flex-col space-y-3 py-2">
        <div>
          <span className="text-xl font-semibold">Добавление серийных номеров</span>
        </div>
        <div>
          <label htmlFor="serialNumberAdd">Добавление </label>
          <div className="flex space-x-2" id="serialNumberAdd">
            <Input name="serialNumber" type="text" value={serialNumber} onChange={(e) => setSerialNumber(e.target.value)}/>
            <Button text="Добавить" buttonType="default" onClick={onAddClick}/>
          </div>
        </div>
        <div>
          <label htmlFor="serialNumbers">Серийные номера</label>
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
                value={selectedSerialNumber}
                options={serialNumbers}
                onChange={(value) => setSelectedSerialNumber({label: value?.label ?? "", value: value?.value ?? ""})}
              />
            </div>
            <Button 
              text="Удалить" 
              buttonType="delete" 
              disabled={selectedSerialNumber.label == "" && selectedSerialNumber.value == ""} 
              onClick={onDeleteClick}
            />
          </div>
        </div>
        <div>
          <Button text="Привязать к материалу" buttonType="default" onClick={() => addSerialNumbersToInvoice(serialNumbers.map<string>(value => value.value))}/>
        </div>
      </div>
    </Modal>
  )
}
