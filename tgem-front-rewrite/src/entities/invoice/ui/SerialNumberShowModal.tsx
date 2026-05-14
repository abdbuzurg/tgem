import { InvoiceMaterialViewWithSerialNumbers } from "@entities/invoice/types/invoiceMaterial";
import Modal from "@shared/components/Modal";

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
  data: InvoiceMaterialViewWithSerialNumbers
}

export default function SerialNumbersShowModal({ setShowModal, data }: Props) {
  return (
    <Modal setShowModal={setShowModal}>
      <div className="py-1 px-2">
        <span className="font-bold">Серийные номера материала - {data.materialName}, {data.costM19}</span>
        <div className="grid grid-cols-2 gap-2">
          {data.serialNumbers.map((value, index) =>
            <div key={index}>{value}</div>
          )}
        </div>
      </div>
    </Modal>
  )
}
