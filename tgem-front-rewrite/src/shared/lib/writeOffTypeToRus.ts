export default function writeOffTypeToRus(writeOffType: string): string {
  switch (writeOffType) {
    case "writeoff-warehouse":
      return "списание со склада"
    case "loss-warehouse":
      return "утеря со склада"
    case "loss-team":
      return "утеря бригады"
    case "loss-object":
      return "утеря с объекта"
    case "writeoff-object":
      return "списание с объекта"

    default: 
      throw new Error("Неправильный тип списания")
  }
}
