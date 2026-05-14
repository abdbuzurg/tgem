import { AxiosResponse } from "axios"
import fileDownload from "js-file-download"
import IApiResponseFormat from "@shared/api/envelope"

// The backend signals errors via the {success, error, ...} envelope at HTTP
// 200 even on file-download endpoints. With axios responseType: "blob" the
// response body is always a Blob, so any envelope-error JSON gets handed to
// fileDownload and saved as a corrupt .xlsx/.pdf. unwrapBlobOrThrow detects
// the JSON case via Content-Type, reads the envelope, and throws on
// success === false; otherwise it returns the file blob.
export async function unwrapBlobOrThrow(response: AxiosResponse<Blob>): Promise<Blob> {
  const contentType = (response.headers["content-type"] ?? "") as string
  if (contentType.includes("application/json")) {
    const text = await response.data.text()
    const envelope = JSON.parse(text) as IApiResponseFormat<unknown>
    if (envelope.success) {
      throw new Error("expected file download but got JSON success envelope")
    }
    throw new Error(envelope.error || "Не удалось загрузить файл")
  }
  return response.data
}

export async function downloadBlob(response: AxiosResponse<Blob>, filename: string): Promise<boolean> {
  const blob = await unwrapBlobOrThrow(response)
  fileDownload(blob, filename)
  return true
}
