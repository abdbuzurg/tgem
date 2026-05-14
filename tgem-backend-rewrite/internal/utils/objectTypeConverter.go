package utils

func ObjectTypeConverter(objectType string) string {
	switch objectType {
	case "tp_objects":
		return "ТП"
	case "kl04kv_objects":
		return "КЛ 04 КВ"
	case "mjd_objects":
		return "МЖД"
	case "sip_objects":
		return "СИП"
	case "stvt_objects":
		return "СТВТ"
	case "substation_objects":
		return "Подстанция"
	default:
		return "Другое"
	}
}
