import Modal from "@shared/components/Modal"
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import { InvoiceOutputOutOfProjectReportFilter, buildReportInvoiceOutputOutOfProject } from "./api";
import { useState } from "react";
import Button from "@shared/ui/Button";
import { useMutation } from "@tanstack/react-query";
import toast from "react-hot-toast";
import LoadingDots from "@shared/ui/LoadingDots";

interface Props {
  setShowReportModal: React.Dispatch<React.SetStateAction<boolean>>
}

export default function ReportInvoiceOutputOutOfProject({ setShowReportModal }: Props) {

  const [filter, setFilter] = useState<InvoiceOutputOutOfProjectReportFilter>({
    dateFrom: null,
    dateTo: null,
  })

  const buildReport = useMutation({
    mutationFn: () => buildReportInvoiceOutputOutOfProject(filter)
  })

  const onCreateReportClick = () => {

    if (filter.dateFrom && filter.dateTo && filter.dateFrom > filter.dateTo) {
      toast.error("Неправильно указан диапазон дат")
      return
    }

    buildReport.mutate()
  }

  return (
    <Modal setShowModal={setShowReportModal}>
      <div className="py-2">
        <p className="text-2xl font-bold">Отсчет накладной отпуск вне проекта</p>
      </div>
      <div className="px-2 flex flex-col space-y-2 pb-2">
        <span className="text-xl font-semibold">Параметры</span>
        <div className="felx flex-col space-y-1">
          <label htmlFor="rangeDate">Диапозон Дат</label>
          <div className="flex space-x-2">
            <div className="py-[4px] px-[8px] border-[#cccccc] border rounded-[4px]">
              <DatePicker
                name="dateOfInvoice"
                className="outline-none w-full"
                dateFormat={"dd-MM-yyyy"}
                selected={filter.dateFrom}
                onChange={(date: Date | null) =>
                  setFilter({ ...filter, dateFrom: date ?? new Date() })
                }
              />
            </div>
            <div className="py-[4px] px-[8px] border-[#cccccc] border rounded-[4px]">
              <DatePicker
                name="dateOfInvoice"
                className="outline-none w-full"
                dateFormat={"dd-MM-yyyy"}
                selected={filter.dateTo}
                onChange={(date: Date | null) =>
                  setFilter({ ...filter, dateTo: date ?? new Date() })
                }
              />
            </div>
            <div>
              <Button
                onClick={() =>
                  setFilter({
                    ...filter,
                    dateFrom: null,
                    dateTo: null
                  })
                }
                text="X"
                buttonType="default"
              />
            </div>
          </div>
        </div>
        <div className="flex">
          <div
            onClick={() => onCreateReportClick()}
            className="text-center text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
          >
            {buildReport.isLoading ? <LoadingDots height={30} /> : "Создать Отчет"}
          </div>
        </div>
      </div>
    </Modal>
  )
}
