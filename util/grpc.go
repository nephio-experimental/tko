package util

func GRPCDefaults(protocol string, address string) (string, string) {
	if protocol == "" {
		protocol = "tcp"
	}

	if address == "" {
		if protocol == "tcp4" {
			address = "0.0.0.0"
		} else {
			address = "::"
		}
	}

	return protocol, address
}
