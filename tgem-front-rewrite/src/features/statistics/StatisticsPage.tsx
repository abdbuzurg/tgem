import { PieChart } from "@mui/x-charts";
import { useQuery } from "@tanstack/react-query";
import { PieChartStat, invoiceCountStat, invoiceInputCreatorStat, invoiceOutputCreatorStat, materialInInvoice, materialInLocation } from "@features/statistics/api";
import LoadingDots from "@shared/ui/LoadingDots";
import { useEffect, useState } from "react";
import Select from 'react-select'
import Material from "@entities/material/types";
import getAllMaterials from "@entities/material/api/getAll";
import IReactSelectOptions from "@shared/types/ReactSelectOptions";

export default function Statistics() {

  const invoiceCountStatQuery = useQuery<PieChartStat[], Error>({
    queryKey: ["invoice-count-stat"],
    queryFn: invoiceCountStat,
  })

  const invoiceInputCreatorStatQuery = useQuery<PieChartStat[], Error>({
    queryKey: ["invoice-input-creator-stat"],
    queryFn: invoiceInputCreatorStat,
  })

  const invoiceOutputCreatorStatQuery = useQuery<PieChartStat[], Error>({
    queryKey: ["invoice-output-creator-stat"],
    queryFn: invoiceOutputCreatorStat,
  })

  const [selectedMaterial, setSelectedMaterial] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [availableMaterials, setAvailableMaterials] = useState<IReactSelectOptions<number>[]>([])
  const allMaterialsQuery = useQuery<Material[], Error>({
    queryKey: ["all-materials"],
    queryFn: () => getAllMaterials(),
  })
  useEffect(() => {
    if (allMaterialsQuery.isSuccess && allMaterialsQuery.data) {
      setAvailableMaterials(allMaterialsQuery.data.map((value) => ({ label: value.name, value: value.id })))
    }
  }, [allMaterialsQuery.data, allMaterialsQuery.isSuccess])

  const materialInInvoicesQuery = useQuery<PieChartStat[], Error>({
    queryKey: ["material-in-invoices", selectedMaterial.value],
    queryFn: () => materialInInvoice(selectedMaterial.value),
    enabled: selectedMaterial.value != 0,
  })

  const materialInLocationQuery = useQuery<PieChartStat[], Error>({
    queryKey: ["material-in-locations", selectedMaterial.value],
    queryFn: () => materialInLocation(selectedMaterial.value),
    enabled: selectedMaterial.value != 0,
  })

  return (
    <main className="px-4">
      <div>
        <p className="text-2xl font-bold">Статистика по проекту</p>
      </div>
      <div className="flex space-x-2">
        <div className="py-2 flex flex-col space-y-2">
          <p className="text-xl font-bold">Соотношение операций</p>
          <div>
            {invoiceCountStatQuery.isLoading && <LoadingDots />}
            {invoiceCountStatQuery.isError &&
              <div>{invoiceCountStatQuery.error.message}</div>
            }
            {invoiceCountStatQuery.isSuccess &&
              <PieChart
                series={[
                  {
                    data: [
                      ...invoiceCountStatQuery.data
                    ],
                  },
                ]}
                width={400}
                height={200}
              />
            }
          </div>
        </div>
        <div className="py-2 flex flex-col space-y-2">
          <p className="text-xl font-bold">Составители Накладной Прихода</p>
          <div>
            {invoiceInputCreatorStatQuery.isLoading && <LoadingDots />}
            {invoiceInputCreatorStatQuery.isError &&
              <div>{invoiceInputCreatorStatQuery.error.message}</div>
            }
            {invoiceInputCreatorStatQuery.isSuccess &&
              <PieChart
                series={[
                  {
                    data: [
                      ...invoiceInputCreatorStatQuery.data
                    ],
                  },
                ]}
                slotProps={{ legend: { hidden: true } }}
                width={400}
                height={200}
              />
            }
          </div>
        </div>
        <div className="py-2 flex flex-col space-y-2">
          <p className="text-xl font-bold">Составители Накладной Отпуск</p>
          <div>
            {invoiceOutputCreatorStatQuery.isLoading && <LoadingDots />}
            {invoiceOutputCreatorStatQuery.isError &&
              <div>{invoiceOutputCreatorStatQuery.error.message}</div>
            }
            {invoiceOutputCreatorStatQuery.isSuccess &&
              <PieChart
                series={[
                  {
                    data: [
                      ...invoiceOutputCreatorStatQuery.data
                    ],
                  },
                ]}
                slotProps={{ legend: { hidden: true } }}
                width={400}
                height={200}
              />
            }
          </div>
        </div>
      </div>
      <div>
        <p className="text-2xl font-bold">Статиска по материалам</p>
        <div>
          <div className="flex flex-col space-y-1">
            <label htmlFor="warehouse-manager">Выберите материал</label>
            <div className="w-[400px]">
              <Select
                className="basic-single"
                classNamePrefix="select"
                isSearchable={true}
                isClearable={true}
                name="warehouse-manager"
                placeholder={""}
                value={selectedMaterial}
                options={availableMaterials}
                onChange={(value) => {
                  setSelectedMaterial(value ?? { label: "", value: 0 })
                }}
              />
            </div>
          </div>
        </div>
        <div className="flex space-x-2">
          {selectedMaterial.value != 0 && <div className="py-2 flex flex-col space-y-2">
            <p className="text-xl font-bold">Количество Материала в операциях</p>
            <div>
              {materialInInvoicesQuery.isLoading && <LoadingDots />}
              {materialInInvoicesQuery.isError &&
                <div>{materialInInvoicesQuery.error.message}</div>
              }
              {materialInInvoicesQuery.isSuccess &&
                <PieChart
                  series={[
                    {
                      data: [
                        ...materialInInvoicesQuery.data
                      ],
                    },
                  ]}
                  slotProps={{ legend: { hidden: true } }}
                  width={400}
                  height={200}
                />
              }
            </div>
          </div>
          }
          {selectedMaterial.value != 0 && <div className="py-2 flex flex-col space-y-2">
            <p className="text-xl font-bold">Количество материала на "складах"</p>
            <div>
              {materialInLocationQuery.isLoading && <LoadingDots />}
              {materialInLocationQuery.isError &&
                <div>{materialInLocationQuery.error.message}</div>
              }
              {materialInLocationQuery.isSuccess &&
                <PieChart
                  series={[
                    {
                      data: [
                        ...materialInLocationQuery.data
                      ],
                    },
                  ]}
                  slotProps={{ legend: { hidden: true } }}
                  width={400}
                  height={200}
                />
              }
            </div>
          </div>
          }
        </div>
      </div>
    </main>
  )
}
