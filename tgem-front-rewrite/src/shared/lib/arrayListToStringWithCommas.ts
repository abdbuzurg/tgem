export default function arrayListToString(array: string[]): string {
  if (array.length == 0) return ""
  
  const result = array.reduce((acc, val) => acc  + ", " + val)

  return result
}
