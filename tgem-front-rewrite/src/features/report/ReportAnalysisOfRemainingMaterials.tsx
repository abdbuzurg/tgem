import { useMutation } from "@tanstack/react-query"
import { analysisOfRemainingMaterials } from "@features/report/mainReportsApi"
import Modal from "@shared/components/Modal"
import LoadingDots from "@shared/ui/LoadingDots"

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
}

export default function AnalysisOfRemainingMaterials({setShowModal}: Props) {

  const analysisOfRemainingMaterialsMutation = useMutation<boolean, Error>({
    mutationFn: analysisOfRemainingMaterials,
  })

  const onCreateReportClick = () => {
    analysisOfRemainingMaterialsMutation.mutate()
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
            {analysisOfRemainingMaterialsMutation.isLoading ? <LoadingDots height={30} /> : "Создать Отчет"}
          </div>
        </div>
      </div>
    </Modal>
  )
}
