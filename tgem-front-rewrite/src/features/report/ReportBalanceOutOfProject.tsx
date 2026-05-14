import { useMutation } from "@tanstack/react-query";
import Modal from "@shared/components/Modal";
import LoadingDots from "@shared/ui/LoadingDots";
import { buildOutOfProjectBalanceReport } from "@features/material-location/api";

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
}

export default function ReportBalanceOutOfProject({ setShowModal }: Props) {

  const buildReportBalanceMutation = useMutation({
    mutationFn: () => buildOutOfProjectBalanceReport()
  })

  const onCreateReportClick = () => {
    buildReportBalanceMutation.mutate()
  }

  return (
    <Modal setShowModal={setShowModal}>
      <div className="py-2">
        <p className="text-2xl font-bold">Отчет остатка материалов вышедших проекта</p>
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
