import ObjectCrudPage from "./_scaffold/ObjectCrudPage"
import { substationConfig } from "./configs/substation"

export default function SubstationObjectPage() {
  return <ObjectCrudPage config={substationConfig} />
}
