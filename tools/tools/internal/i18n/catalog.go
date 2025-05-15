package i18n

var en = map[string]string{
	"delete.dup":      "Duplicate hostname: %s (total: %d)",
	"delete.keep":     "  Keep: machine_id=%s date=%s",
	"delete.drop":     "  Delete: machine_id=%s date=%s",
	"delete.unknown":  "unknown",
	"delete.hosts":    "Hostnames with duplicates: %d",
	"delete.dry":      "Records marked (dry-run): %d",
	"delete.removed":  "Records removed: %d",
	"delete.err":      "error removing %s: %v",
	"update.loaded":   "Total numbers loaded: %d",
	"update.empty":    "No valid numbers found in the report. Nothing to do.",
	"update.hosts":    "Hostnames analyzed: %d",
	"update.numbered": "Hostnames with an identified number: %d",
	"update.matches":  "Matches found: %d",
	"update.updated":  "Records updated: %d",
	"update.err":      "error updating %s: %v",
	"add.hostname":    "--hostname is required",
}

var pt = map[string]string{
	"delete.dup":      "Hostname duplicado: %s (total: %d)",
	"delete.keep":     "  Manter: machine_id=%s data=%s",
	"delete.drop":     "  Apagar: machine_id=%s data=%s",
	"delete.unknown":  "desconhecida",
	"delete.hosts":    "Hostnames com duplicatas: %d",
	"delete.dry":      "Registros marcados (dry-run): %d",
	"delete.removed":  "Registros removidos: %d",
	"delete.err":      "erro ao remover %s: %v",
	"update.loaded":   "Total de números carregados: %d",
	"update.empty":    "Nenhum número válido encontrado no relatório. Nada a fazer.",
	"update.hosts":    "Hostnames analisados: %d",
	"update.numbered": "Hostnames com número identificado: %d",
	"update.matches":  "Correspondências encontradas: %d",
	"update.updated":  "Registros atualizados: %d",
	"update.err":      "erro ao atualizar %s: %v",
	"add.hostname":    "--hostname é obrigatório",
}
