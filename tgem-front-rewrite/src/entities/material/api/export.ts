import fileDownload from "js-file-download";
import Material from "@entities/material/types";
import axiosClient from "@shared/api/client";

export default async function exportMaterials(data: Material[]): Promise<boolean> {
  const responseRaw = await axiosClient.post("/material/export", data, { responseType: "blob" })
  if (responseRaw.status == 200) {
    fileDownload(responseRaw.data, "Материалы.xlsx")
    return true
  } else {
    throw new Error(responseRaw.data)
  }
}