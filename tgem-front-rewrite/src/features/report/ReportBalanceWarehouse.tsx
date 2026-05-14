import { useMutation } from "@tanstack/react-query"
import { useState } from "react"
import { ReportBalanceFilter, buildReportBalance } from "@features/report/api"
import LoadingDots from "@shared/ui/LoadingDots"
import Modal from "@shared/components/Modal"

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
}

export default function ReportBalanceWarehouse({ setShowModal }: Props) {
  const [filter] = useState<ReportBalanceFilter>({
    objectID: 0,
    teamID: 0,
    type: "warehouse",
  })

  //Submit filter
  const buildReportBalanceMutation = useMutation({
    mutationFn: () => buildReportBalance(filter)
  })

  const onCreateReportClick = () => {
    buildReportBalanceMutation.mutate()
  }

  return (
    <Modal setShowModal={setShowModal}>
      <div className="pb-2">
        <p className="text-2xl font-bold">Отсчет остатка склада</p>
      </div>
      <div className="px-2 flex flex-col space-y-2 pb-2">
        <div className="flex">
          <div
            onClick={() => onCreateReportClick()}
            className="text-center text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
          >
            {buildReportBalanceMutation.isLoading ? <LoadingDots height={30} /> : "Создать Отчет"}
          </div>
        </div>
      </div>
    </Modal>
  )
}
