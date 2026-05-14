import { useState } from "react"
import AnalysisOfRemainingMaterials from "@features/report/ReportAnalysisOfRemainingMaterials"
import ReportProjectProgress from "@features/report/ReportProjectProgress"

const mainReport = [
  {
    title: "Прогресс Реализации Проекта",
  },
  {
    title: "Аналитика Остатка Материалов",
  }
]


export default function MainReports() {
  const [mainReportTitle, setMainReportTitle] = useState("")
  const [showMainReportModal, setShowMainReportModal] = useState(false)

  const getReport = (title: string) => {
    setShowMainReportModal(true)
    setMainReportTitle(title)
  }

  return (
    <div className="py-1 px-2">
      <span className="font-bold text-2xl">Отсчетность</span>
      <div className="grid grid-cols-4 gap-2 px-2 py-2">
        {mainReport.map((val, index) => (
          <div
            key={index}
            className="bg-gray-800 text-white text-xl py-3 px-6 rounded cursor-pointer hover:bg-gray-900 text-center align-middle"
            onClick={() => getReport(val.title)}
          >{val.title}</div>
        ))}
      </div>
      {(showMainReportModal && (
        (mainReportTitle == "Прогресс Реализации Проекта" && <ReportProjectProgress setShowModal={setShowMainReportModal}/>)
        || (mainReportTitle == "Аналитика Остатка Материалов" && <AnalysisOfRemainingMaterials setShowModal={setShowMainReportModal} />)
      ))}
    </div>
  )
}
