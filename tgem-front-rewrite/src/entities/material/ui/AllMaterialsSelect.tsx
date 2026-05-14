import Select from 'react-select'
import IReactSelectOptions from '@shared/types/ReactSelectOptions'
import { useState } from 'react'
import { IMaterialCost } from '@entities/material-cost/types'
import useMaterialOptions from '../hooks/useMaterialOptions'

interface Props {
  valueName: string
  setValueDispatcher: React.Dispatch<React.SetStateAction<IMaterialCost>>
}

export default function AllMaterialsSelect({ valueName, setValueDispatcher }: Props) {
  const options = useMaterialOptions()
  const [selectedData, setSelectedData] = useState<IReactSelectOptions<number>>({ value: 0, label: "" })

  const onSelectChange = (value: IReactSelectOptions<number> | null) => {
    if (!value) {
      setSelectedData({ label: "", value: 0 })
      setValueDispatcher((prev) => ({ ...prev, [valueName]: 0 }))
      return
    }
    setValueDispatcher((prev) => ({ ...prev, [valueName]: value.value }))
    setSelectedData(value)
  }

  return (
    <Select
      className="basic-single"
      classNamePrefix="select"
      isSearchable={true}
      isClearable={true}
      name={"materials"}
      placeholder={""}
      value={selectedData}
      options={options}
      onChange={(value) => onSelectChange(value)}
    />
  )
}
