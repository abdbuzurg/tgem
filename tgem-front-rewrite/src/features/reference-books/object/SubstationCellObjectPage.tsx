import ObjectCrudPage from "./_scaffold/ObjectCrudPage"
import { substationCellConfig } from "./configs/substation-cell"

export default function SubstationCellObjectPage() {
  return <ObjectCrudPage config={substationCellConfig} />
}
