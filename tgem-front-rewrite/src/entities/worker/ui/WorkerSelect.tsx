import Select from 'react-select'
import IReactSelectOptions from '@shared/types/ReactSelectOptions'
import useWorkerOptions from '../hooks/useWorkerOptions'

interface Props {
  title: string
  jobTitle: string
  selectedWorkerID: IReactSelectOptions<number>
  setSelectedWorkerID: React.Dispatch<React.SetStateAction<IReactSelectOptions<number>>>
  error?: boolean
}

export default function WorkerSelect({ title, jobTitle, selectedWorkerID, setSelectedWorkerID, error }: Props) {
  const options = useWorkerOptions(jobTitle)

  return (
    <div className="flex flex-col space-y-1">
      <label htmlFor={jobTitle}>{title[0].toUpperCase() + title.substring(1, title.length)}</label>
      <div className="w-[200px]">
        <Select
          className="basic-single"
          classNamePrefix="select"
          isSearchable={true}
          isClearable={true}
          name={jobTitle}
          placeholder={""}
          value={selectedWorkerID}
          options={options}
          onChange={(value) => setSelectedWorkerID(value ?? { label: "", value: 0 })}
        />
      </div>
      {error && <p className="text-red-500 text-sm font-semibold">Не выбрано</p>}
    </div>
  )
}
