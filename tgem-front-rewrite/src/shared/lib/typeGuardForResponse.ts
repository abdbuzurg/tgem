import IApiResponseFormat from "@shared/api/envelope";

export default function isCorrectResponseFormat<T>(object: unknown): object is IApiResponseFormat<T> {
  if (typeof object !== "object" || object === null) return false
  return 'data' in object && 'success' in object && 'permission' in object && 'error' in object
}
