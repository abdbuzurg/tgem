import { useMutation } from "@tanstack/react-query"
import { projectProgressReport } from "@features/report/mainReportsApi"
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import Modal from "@shared/components/Modal"
import LoadingDots from "@shared/ui/LoadingDots"
import { useState } from "react";

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
}

export default function ReportProjectProgress({ setShowModal }: Props) {

  const [date, setDate] = useState(new Date())

  const projectProgressReportMutation = useMutation<boolean, Error, Date>({
    mutationFn: projectProgressReport,
  })

  const onCreateReportClick = () => {
    projectProgressReportMutation.mutate(date)
  }

  return (
    <Modal setShowModal={setShowModal}>
      <div className="pb-2">
        <p className="text-2xl font-bold">Отчет - Прогресс реализации проекта</p>
      </div>
      <div className="px-2 flex flex-col space-y-2 pb-2">
        <div className="felx flex-col space-y-1">
          <label htmlFor="rangeDate">Дата</label>
          <div className="flex space-x-2">
            <div className="py-[4px] px-[8px] border-[#cccccc] border rounded-[4px]">
              <DatePicker
                name="progressReportDate"
                className="outline-none w-full"
                dateFormat={"dd-MM-yyyy"}
                selected={date}
                maxDate={new Date()}
                onChange={(date: Date | null) => setDate(date ?? new Date())}
              />
            </div>
          </div>
          <span></span>
        </div>
        <div className="flex">
          <div
            onClick={() => onCreateReportClick()}
            className="text-center text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
          >
            {projectProgressReportMutation.isLoading ? <LoadingDots height={30} /> : "Создать Отчет"}
          </div>
        </div>
      </div>
    </Modal>
  )
}
