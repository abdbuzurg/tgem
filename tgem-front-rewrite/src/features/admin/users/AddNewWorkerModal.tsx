import Modal from "@shared/components/Modal"
import Input from "@shared/ui/Input"
import Select from 'react-select'
import Button from "@shared/ui/Button"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import toast from "react-hot-toast"
import IWorker from "@entities/worker/types"
import { useState } from "react"
import { JOB_TITLES } from "@shared/lib/data/jobTitles"
import { createWorker } from "@entities/worker/api"
import LoadingDots from "@shared/ui/LoadingDots"

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
}

export default function AddNewWorkerModal({ setShowModal }: Props) {
  const [workerMutationData, setWorkerMutationData] = useState<IWorker>({
    id: 0,
    jobTitleInCompany: "",
    jobTitleInProject: "",
    companyWorkerID: "",
    mobileNumber: "",
    name: "",
  })
  //Submitting the worker repo
  const queryClient = useQueryClient()
  const [isSelectedFromList, setIsSelectedForList] = useState(false)

  const createWorkerMutation = useMutation<IWorker, Error, IWorker>({
    mutationFn: createWorker,
    onSettled: () => {
      queryClient.invalidateQueries(["workers"])
      setShowModal(false)
    }
  })

  const onMutationSubmit = () => {
    if (workerMutationData.name == "") {
      toast.error("Не указано имя работника")
      return
    }

    if (workerMutationData.jobTitleInCompany == "") {
      toast.error("Не указано должность рабоника в ТГЭМ")
      return
    }

    if (workerMutationData.jobTitleInProject == "") {
      toast.error("Не указано должность рабоника в проекте")
      return
    }

    const loadingToast = toast.loading("Сохранение новых данных...")
    createWorkerMutation.mutate(workerMutationData, {
      onSuccess: () => {
        toast.success("Сохранение прошло успешно")
        queryClient.invalidateQueries(["all-workers"])
        setShowModal(false)
      },
      onSettled: () => {
        toast.dismiss(loadingToast)
      }
    })
  }

  return (
    <Modal setShowModal={setShowModal}>
      <div className="">
        <h3 className="text-xl font-medium text-gray-800">
          Добавление сотрудника
        </h3>
        <div className="flex flex-col space-y-3 mt-2">
          <div className="flex flex-col space-y-1">
            <label htmlFor="name">Фамилия Имя Отчество<span className="text-red-600">*</span></label>
            <Input
              name="name"
              type="text"
              value={workerMutationData.name}
              onChange={(e) => setWorkerMutationData({ ...workerMutationData, [e.target.name]: e.target.value })}
            />
          </div>
          <div className="flex flex-col space-y-1">
            <label htmlFor="jobTitle">Должность в ТГЭМ<span className="text-red-600">*</span></label>
            <Select
              className="basic-single"
              classNamePrefix="select"
              isSearchable={true}
              isClearable={true}
              name={"job-title-in-company"}
              placeholder={""}
              value={{ label: workerMutationData.jobTitleInCompany, value: workerMutationData.jobTitleInCompany }}
              options={JOB_TITLES.map<IReactSelectOptions<string>>((value) => ({ label: value, value: value }))}
              onChange={(value) => {
                setWorkerMutationData({
                  ...workerMutationData,
                  jobTitleInCompany: value?.value ?? "",
                })
              }}
            />
          </div>
          <div className="flex flex-col space-y-1">
            <label htmlFor="mobileNumber">ID сотрудника<span className="text-red-600">*</span></label>
            <Input
              name="companyWorkerID"
              type="text"
              value={workerMutationData.companyWorkerID}
              onChange={(e) => setWorkerMutationData({ ...workerMutationData, [e.target.name]: e.target.value })}
            />
          </div>
          <div className="flex flex-col space-y-1">
            <label htmlFor="jobTitle">Должность в Проекте<span className="text-red-600">*</span></label>
            <div className="flex w-full space-x-2">
              <Select
                className="basic-single flex-1"
                classNamePrefix="select"
                isSearchable={true}
                isClearable={true}
                name={"job-title-in-project"}
                placeholder={""}
                value={{ label: workerMutationData.jobTitleInProject, value: workerMutationData.jobTitleInCompany }}
                options={JOB_TITLES.map<IReactSelectOptions<string>>((value) => ({ label: value, value: value }))}
                onChange={(value) => {
                  setWorkerMutationData({
                    ...workerMutationData,
                    jobTitleInProject: value?.value ?? "",
                  })
                  if (!value) {
                    setIsSelectedForList(false)
                  } else {
                    setIsSelectedForList(true)
                  }
                }}
              />
              <div className="flex-1">
                <Input
                  name="jobTitleInProject"
                  type="text"
                  value={workerMutationData.jobTitleInProject}
                  onChange={(e) => {
                    if (isSelectedFromList) {
                      return
                    }

                    setWorkerMutationData({
                      ...workerMutationData,
                      jobTitleInProject: e.target.value,
                    })
                  }}
                />
              </div>
            </div>
          </div>
          <div className="flex flex-col space-y-1">
            <label htmlFor="mobileNumber">Номера телефона<span className="text-red-600">*</span></label>
            <Input
              name="mobileNumber"
              type="text"
              value={workerMutationData.mobileNumber}
              onChange={(e) => setWorkerMutationData({ ...workerMutationData, [e.target.name]: e.target.value })}
            />
          </div>
          <div>
            {createWorkerMutation.isLoading
              ?
              <LoadingDots height={30} />
              :
              <Button
                text="Добавить"
                onClick={onMutationSubmit}
              />
            }
          </div>
        </div>
      </div>
    </Modal>
  )
}
