package plugins

func IsValidPluginType(type_ string, allowEmpty bool) bool {
	switch type_ {
	case "validate", "prepare", "schedule":
		return true
	case "":
		return allowEmpty
	default:
		return false
	}
}

const PluginTypesDescription = "\"validate\", \"prepare\", or \"schedule\""
