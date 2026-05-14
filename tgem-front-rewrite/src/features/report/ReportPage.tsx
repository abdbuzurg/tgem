import { useState } from "react"
import { REPORT_BALANCE_LOSS_OBJECT, REPORT_BALANCE_LOSS_TEAM, REPORT_BALANCE_LOSS_WAREHOUSE, REPORT_BALANCE_OBJECT, REPORT_BALANCE_OUT_OF_PROJECT, REPORT_BALANCE_TEAM, REPORT_BALANCE_WAREHOUSE, REPORT_BALANCE_WRITEOFF_OBJECT, REPORT_BALANCE_WRITEOFF_WAREHOUSE, REPORT_HR_ATTENDANCE, REPORT_INVOICE_INPUT, REPORT_INVOICE_LOSS_OBJECT, REPORT_INVOICE_LOSS_TEAM, REPORT_INVOICE_LOSS_WAREHOUSE, REPORT_INVOICE_OBJECT, REPORT_INVOICE_OUTPUT, REPORT_INVOICE_OUTPUT_OUT_OF_PROJECT, REPORT_INVOICE_RETURN, REPORT_INVOICE_WRITEOFF_OBJECT, REPORT_INVOICE_WRITEOFF_WAREHOUSE } from "@routes/paths"
import ReportInvoiceInput from "@features/invoice/input/ReportInvoiceInput"
import ReportInvoiceOutput from "@features/invoice/output-in-project/ReportInvoiceOutput"
import ReportInvoiceReturn from "@entities/invoice-return/ui/ReportInvoiceReturn"
import ReportInvoiceObject from "@features/invoice/object/ReportInvoiceObject"
import ReportBalanceWarehouse from "./ReportBalanceWarehouse"
import ReportBalanceTeam from "./ReportBalanceTeam"
import ReportBalanceObject from "./ReportBalanceObject"
import ReportBalanceWriteOffWarehouse from "./ReportBalanceWriteOffWarehouse"
import ReportBalanceLossWarehouse from "./ReportBalanceLossWarehouse"
import ReportBalanceLossTeam from "./ReportBalanceLossTeam"
import ReportBalanceLossObject from "./ReportBalanceLossObject"
import ReportBalanceWriteOffObject from "./ReportBalanceWriteOffObject"
import ReportInvoiceWriteOff from "@entities/invoice-writeoff/ui/ReportInvoiceWriteOff"
import ReportBalanceOutOfProject from "./ReportBalanceOutOfProject"
import ReportInvoiceOutputOutOfProject from "@features/invoice/output-out-of-project/ReportInvoiceOutputOutOfProject"

const report = [
  {
    category: "Остаток",
    pages: [
      { name: "Материалы на складе", type: REPORT_BALANCE_WAREHOUSE },
      { name: "Материалы у бригад", type: REPORT_BALANCE_TEAM },
      { name: "Материалы на oбъектах", type: REPORT_BALANCE_OBJECT },
      { name: "Материалы вышедшие с проекта", type: REPORT_BALANCE_OUT_OF_PROJECT },
      { name: "Материалы списанные со склада", type: REPORT_BALANCE_WRITEOFF_WAREHOUSE },
      { name: "Утеренные материалы со склада", type: REPORT_BALANCE_LOSS_WAREHOUSE },
      { name: "Утеренные материалы бригад", type: REPORT_BALANCE_LOSS_TEAM },
      { name: "Утеренные материалы объекта", type: REPORT_BALANCE_LOSS_OBJECT },
      { name: "Материалы списанные с объекта", type: REPORT_BALANCE_WRITEOFF_OBJECT },
    ]
  },
  {
    category: "Операции",
    pages: [
      { name: "Отчет приход", type: REPORT_INVOICE_INPUT },
      { name: "Отчет отпуск внутри проекта", type: REPORT_INVOICE_OUTPUT },
      { name: "Отчет отпуск вне проекта", type: REPORT_INVOICE_OUTPUT_OUT_OF_PROJECT },
      { name: "Отчет возврат", type: REPORT_INVOICE_RETURN },
      { name: "Отчет расхода на объект", type: REPORT_INVOICE_OBJECT },
      { name: "Отчет списания со склада", type: REPORT_INVOICE_WRITEOFF_WAREHOUSE },
      { name: "Отчет утери со склада", type: REPORT_INVOICE_LOSS_WAREHOUSE },
      { name: "Отчет утери бригады", type: REPORT_INVOICE_LOSS_TEAM },
      { name: "Отчет утери с объекта", type: REPORT_INVOICE_LOSS_OBJECT },
      { name: "Отчет списания с объекта", type: REPORT_INVOICE_WRITEOFF_OBJECT },
    ]
  },
  {
    category: "Кадр",
    pages: [
      { name: "Отчет по МОРФО", type: REPORT_HR_ATTENDANCE },
    ]
  },
]

export default function Report() {

  const [selectedReportModal, setSelectedReportModal] = useState(false)
  const [selectedReportType, setSelectedReportType] = useState("")

  const showReportParameters = (type: string) => {
    setSelectedReportModal(true)
    setSelectedReportType(type)
  }

  return (
    <div className="flex flex-col space-y-2">
      <div className="px-2">
        <p className="text-3xl font-bold">Отчёт</p>
        {report.map((val, index) => (
          <div className="py-1 px-2" key={index}>
            <span className="font-bold text-2xl">{val.category}</span>
            <div className="grid grid-cols-4 gap-2 px-2 py-2">
              {val.pages.map((page, index) =>
                <div
                  key={index}
                  className="bg-gray-800 text-white text-xl py-3 px-6 rounded cursor-pointer hover:bg-gray-900 text-center align-middle"
                  onClick={() => showReportParameters(page.type)}
                >
                  {page.name}
                </div>
              )}
            </div>
          </div>
        ))}
      </div>
      {selectedReportModal && (
        (selectedReportType == REPORT_BALANCE_WAREHOUSE && <ReportBalanceWarehouse setShowModal={setSelectedReportModal} />)
        || (selectedReportType == REPORT_BALANCE_TEAM && <ReportBalanceTeam setShowModal={setSelectedReportModal} />)
        || (selectedReportType == REPORT_BALANCE_OBJECT && <ReportBalanceObject setShowModal={setSelectedReportModal} />)
        || (selectedReportType == REPORT_BALANCE_OUT_OF_PROJECT && <ReportBalanceOutOfProject setShowModal={setSelectedReportModal} />)
        || (selectedReportType == REPORT_BALANCE_WRITEOFF_WAREHOUSE && <ReportBalanceWriteOffWarehouse setShowModal={setSelectedReportModal} />)
        || (selectedReportType == REPORT_BALANCE_LOSS_WAREHOUSE && <ReportBalanceLossWarehouse setShowModal={setSelectedReportModal} />)
        || (selectedReportType == REPORT_BALANCE_LOSS_TEAM && <ReportBalanceLossTeam setShowModal={setSelectedReportModal} />)
        || (selectedReportType == REPORT_BALANCE_LOSS_OBJECT && <ReportBalanceLossObject setShowModal={setSelectedReportModal} />)
        || (selectedReportType == REPORT_BALANCE_WRITEOFF_OBJECT && <ReportBalanceWriteOffObject setShowModal={setSelectedReportModal} />)
        || (selectedReportType == REPORT_INVOICE_WRITEOFF_WAREHOUSE && <ReportInvoiceWriteOff writeOffType="writeoff-warehouse" setShowReportModal={setSelectedReportModal} />)
        || (selectedReportType == REPORT_INVOICE_LOSS_WAREHOUSE && <ReportInvoiceWriteOff writeOffType="loss-warehouse" setShowReportModal={setSelectedReportModal} />)
        || (selectedReportType == REPORT_INVOICE_LOSS_TEAM && <ReportInvoiceWriteOff writeOffType="loss-team" setShowReportModal={setSelectedReportModal} />)
        || (selectedReportType == REPORT_INVOICE_LOSS_OBJECT && <ReportInvoiceWriteOff writeOffType="loss-object" setShowReportModal={setSelectedReportModal} />)
        || (selectedReportType == REPORT_INVOICE_WRITEOFF_OBJECT && <ReportInvoiceWriteOff writeOffType="writeoff-object" setShowReportModal={setSelectedReportModal} />)
        || (selectedReportType == REPORT_INVOICE_INPUT && <ReportInvoiceInput setShowReportModal={setSelectedReportModal} />)
        || (selectedReportType == REPORT_INVOICE_OUTPUT && <ReportInvoiceOutput setShowReportModal={setSelectedReportModal} />)
        || (selectedReportType == REPORT_INVOICE_RETURN && <ReportInvoiceReturn setShowReportModal={setSelectedReportModal} />)
        || (selectedReportType == REPORT_INVOICE_OBJECT && <ReportInvoiceObject setShowReportModal={setSelectedReportModal} />)
        || (selectedReportType == REPORT_INVOICE_OUTPUT_OUT_OF_PROJECT && <ReportInvoiceOutputOutOfProject setShowReportModal={setSelectedReportModal} />)
      )
      }
    </div>
  )
}
