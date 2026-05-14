import { useEffect, useState } from "react"
import { MaterialLocationLiveSearchParameters, MaterialLocationLiveView, getMaterialLocation } from "@features/material-location/api"
import { useQuery } from "@tanstack/react-query"
import LoadingDots from "@shared/ui/LoadingDots"
import Material from "@entities/material/types"
import getAllMaterials from "@entities/material/api/getAll"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import Modal from "@shared/components/Modal"
import Select from "react-select"

export default function MaterialLocationLive() {

  const [searchParameters, setSearchParameters] = useState<MaterialLocationLiveSearchParameters>({
    locationType: "warehouse",
    locationID: 0,
    materialID: 0,
  })

  const [tableData, setTableData] = useState<MaterialLocationLiveView[]>([])
  const tableDataQuery = useQuery<MaterialLocationLiveView[], Error, MaterialLocationLiveView[]>({
    queryKey: ["material-location-live", searchParameters],
    queryFn: () => getMaterialLocation(searchParameters)
  })
  useEffect(() => {
    if (tableDataQuery.isSuccess && tableDataQuery.data) {
      setTableData(tableDataQuery.data)
    }
  }, [tableDataQuery.data, tableDataQuery.isSuccess])

  const [showWarehouseSearchModal, setShowWarehouseSearchModal] = useState(false)
  const [selectedMaterial, setSelectedMaterial] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [allMaterials, setAllMaterials] = useState<IReactSelectOptions<number>[]>([])
  const allMaterialsQuery = useQuery<Material[], Error, Material[]>({
    queryKey: ["all-materials"],
    queryFn: getAllMaterials,
    enabled: showWarehouseSearchModal,
  })
  useEffect(() => {
    if (allMaterialsQuery.isSuccess && allMaterialsQuery.data) {
      setAllMaterials(allMaterialsQuery.data.map<IReactSelectOptions<number>>((val) => ({
        value: val.id,
        label: val.name,
      })))
    }
  }, [allMaterialsQuery.data, allMaterialsQuery.isSuccess])

  return (
    <main>
      <div className="mt-2 px-2 flex space-x-3">
        <span className="text-3xl font-bold">Местоположение материала</span>
        <div
          className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
          onClick={() => setShowWarehouseSearchModal(true)}
        >
          Склад
        </div>
        <div className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer">
          Бригады
        </div>
        <div className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer">
          Объект
        </div>
      </div>
      <table className="table-auto text-sm text-left mt-2 w-full border-box">
        <thead className="shadow-md border-t-2">
          <tr>
            <th className="px-4 py-3 w-2/3">
              <span>Наименование</span>
            </th>
            <th className="px-4 py-3 w-[100px]">
              <span>Ед.Изм.</span>
            </th>
            <th className="px-4 py-3">
              <span>Количество</span>
            </th>
            <th className="px-4 py-3">
              <span>Цена</span>
            </th>
            <th className="px-4 py-3">
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
          {tableDataQuery.isSuccess && tableData.map((value, index) => (
            <tr key={index} className="border-b hover:bg-gray-200">
              <td className="px-4 py-3">{value.materialName}</td>
              <td className="px-4 py-3">{value.materialUnit}</td>
              <td className="px-4 py-3">{value.amount}</td>
              <td className="px-4 py-3">{value.materialCostM19}</td>
              <td className="px-4 py-3"></td>
            </tr>
          ))}
        </tbody>
      </table>
      {showWarehouseSearchModal &&
        <Modal setShowModal={setShowWarehouseSearchModal}>
          <span className="font-bold text-xl py-1">Параметры Поиска по материалам в складе</span>
          {allMaterialsQuery.isLoading && <LoadingDots />}
          {allMaterialsQuery.isSuccess &&
            <div className="p-2 flex flex-col space-y-2">
              <div className="flex flex-col space-y-1">
                <label htmlFor="material-names">Материалы</label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"material-names"}
                  placeholder={""}
                  value={selectedMaterial}
                  options={allMaterials}
                  onChange={value => {
                    setSelectedMaterial(value ?? { label: "", value: 0 })
                    setSearchParameters({
                      ...searchParameters,
                      materialID: value?.value ?? 0,
                    })
                  }}
                />
              </div>
            </div>
          }
        </Modal>
      }
    </main>
  )
}
