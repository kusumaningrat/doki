package helper

import (
	"docker-tui/internal/domain"
	"fmt"
	"strings"
)

func FormatContainerPorts(ports []domain.Port) string {
	if len(ports) == 0 {
		return ""
	}

	var result string
	for _, p := range ports {
		if p.PublicPort != 0 {
			result += fmt.Sprintf("%d:%d/%s ", p.PrivatePort, p.PublicPort, p.Type)
		} else {
			result += fmt.Sprintf("%d/%s ", p.PrivatePort, p.Type)
		}
	}

	return strings.TrimSpace(result)
}
