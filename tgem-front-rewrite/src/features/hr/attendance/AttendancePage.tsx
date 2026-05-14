import { useEffect, useState } from "react";
import Button from "@shared/ui/Button";
import LoadingDots from "@shared/ui/LoadingDots";
import Modal from "@shared/components/Modal";
import { useInfiniteQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import getPaginatedWorkerAttendance, { WorkerAttendancePaginated, WorkerAttendancePaginatedResponse, createWorkerAttendance } from "@features/hr/attendance/api";
import toast from "react-hot-toast";
import { ENTRY_LIMIT } from "@shared/config/pagination";
import useScrollPaginated from "@shared/hooks/useScrollPaginated"

export default function Attendance() {

  const queryClient = useQueryClient()

  const tableDataQuery = useInfiniteQuery<WorkerAttendancePaginatedResponse, Error>({
    queryKey: ["worker-attendance"],
    queryFn: ({ pageParam }) => getPaginatedWorkerAttendance({ pageParam }),
    getNextPageParam: (lastPage) => {
      if (lastPage.page * ENTRY_LIMIT > lastPage.count) return undefined
      return lastPage.page + 1
    }
  })
  const [tableData, setTableData] = useState<WorkerAttendancePaginated[]>([])
  useEffect(() => {
    if (tableDataQuery.isSuccess && tableDataQuery.data) {
      const data: WorkerAttendancePaginated[] = tableDataQuery.data.pages.reduce<WorkerAttendancePaginated[]>((acc, page) => [...acc, ...page.data], [])
      setTableData(data)
    }
  }, [tableDataQuery.data, tableDataQuery.isSuccess])

  useScrollPaginated(tableDataQuery.fetchNextPage)

  //mutation data
  const [showAddModal, setShowAddModal] = useState(false)
  const createWorkerAttendanceMutation = useMutation<boolean, Error, File>({
    mutationFn: createWorkerAttendance,
  })

  const acceptAddExcel = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return
    createWorkerAttendanceMutation.mutate(e.target.files[0], {
      onSuccess: () => {
        queryClient.invalidateQueries(["worker-attendance"])
        setShowAddModal(false)
        toast.success("Данные добавлены успешно")
      },
      onSettled: () => {
        e.target.value = ""
      },
      onError: (error) => {
        toast.error(`Импортированный файл имеет неправильные данные: ${error.message}`)
      }
    })
  }

  const extractTime = (dateTimeStr: string): string => {
    const dateTime = new Date(dateTimeStr) 
    let hour = dateTime.getUTCHours().toString()
    if (+hour < 10) hour = "0" + hour
    let minute = dateTime.getUTCMinutes().toString()
    if (+minute < 10) minute = "0" + minute
    let second = dateTime.getUTCSeconds().toString()
    if (+second < 10) second = "0" + second
    return hour + ":" + minute + ":" + second
  }

  const extractDate = (dateTimeStr: string): string => {
    const dateTime = new Date(dateTimeStr)
    let day = dateTime.getUTCDate().toString()
    if (+day < 10) day = "0" + day
    let month = (dateTime.getUTCMonth() + 1).toString()
    if (+month < 10) month = "0" + month
    const year = dateTime.getUTCFullYear()
    return day + "-" + month + "-" + year
  }

  return (
    <main>
      <div className="mt-2 px-2 flex justify-between">
        <span className="text-3xl font-bold">МОРФО сотрудников</span>
      </div>
      <table className="table-auto text-sm text-left mt-2 w-full border-box">
        <thead className="shadow-md border-t-2">
          <tr>
            <th className="px-4 py-3">
              <span>Имя</span>
            </th>
            <th className="px-4 py-3 w-[130px]">
              <span>ID сотрудника</span>
            </th>
            <th className="px-4 py-3">
              <span>Начало дня</span>
            </th>
            <th className="px-4 py-3">
              <span>Конец дня</span>
            </th>
            <th className="px-4 py-3 w-[150px]">
              <span>Дата</span>
            </th>
            <th className="px-4 py-3">
              <Button text="Добавить" onClick={() => setShowAddModal(true)} />
            </th>
          </tr>
        </thead>
        <tbody>
          {tableDataQuery.isLoading &&
            <tr>
              <td colSpan={6}>
                <LoadingDots />
              </td>
            </tr>
          }
          {tableDataQuery.isError &&
            <tr>
              <td colSpan={6} className="text-red font-bold text-center">
                {tableDataQuery.error.message}
              </td>
            </tr>
          }
          {tableDataQuery.isSuccess &&
            tableData.map((row, index) => (
              <tr key={index} className="border-b">
                <td className="px-4 py-3">{row.workerName}</td>
                <td className="px-4 py-3">{row.companyWorkerID}</td>
                <td className="px-4 py-3">{extractTime(row.start)}</td>
                <td className="px-4 py-3">{extractTime(row.end)}</td>
                <td className="px-4 py-3">{extractDate(row.start)}</td>
                <td className="px-4 py-3 border-box flex space-x-3"></td>
              </tr>
            ))
          }
        </tbody>
      </table>
      {showAddModal &&
        <Modal setShowModal={setShowAddModal}>
          <span className="font-bold text-xl px-2 py-1">Импорт данных в Справочник - Материалы</span>
          <div className="grid grid-cols-2 gap-2 items-center px-2 pt-2 h-12">
            <div className="w-full">
              {createWorkerAttendanceMutation.status == "loading"
                ?
                <div className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800">
                  <LoadingDots height={25} />
                </div>
                :
                <label
                  htmlFor="file"
                  className="w-full text-white py-3 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer text-center"
                >
                  Импортировать данные
                </label>
              }
              <input
                name="file"
                type="file"
                id="file"
                onChange={(e) => acceptAddExcel(e)}
                className="hidden"
              />
            </div>
            <div className="w-full invisible">
            </div>
          </div>
          <span className="text-sm italic px-2 w-full text-center">При импортировке система будет следовать правилам шаблона</span>
        </Modal>
      }
    </main>
  )
}
