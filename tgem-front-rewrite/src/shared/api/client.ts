import axios from 'axios';
import IApiResponseFormat from "./envelope";

const axiosClient = axios.create();

axiosClient.defaults.baseURL = import.meta.env.VITE_API_BASE_URL;

axiosClient.defaults.timeout = 60000;

axiosClient.interceptors.request.use(function (config) {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config;
});

// File-download endpoints request responseType: "blob", but the backend
// signals errors via the {success, error, ...} envelope at HTTP 200. Without
// this interceptor, that error JSON gets handed to fileDownload and saved
// as a corrupt .xlsx/.pdf. Detect the JSON-envelope case via Content-Type,
// parse it, and throw so the caller's normal error path runs.
axiosClient.interceptors.response.use(async function (response) {
  if (response.config.responseType === "blob" && response.data instanceof Blob) {
    const contentType = (response.headers["content-type"] ?? "") as string
    if (contentType.includes("application/json")) {
      const text = await response.data.text()
      try {
        const envelope = JSON.parse(text) as IApiResponseFormat<unknown>
        if (envelope && envelope.success === false) {
          throw new Error(envelope.error || "Не удалось загрузить файл")
        }
      } catch (e) {
        if (e instanceof SyntaxError) {
          // Not actually JSON despite the header — let the caller handle it.
          return response
        }
        throw e
      }
    }
  }
  return response
});

export default axiosClient
