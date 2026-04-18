package monitor

import (
	"database/sql"
	"strings"

	sshclient "backend/internal/ssh"
)

// ServiceInfo represents a systemd service's state.
type ServiceInfo struct {
	Name        string `json:"name"`
	LoadState   string `json:"load_state"`
	ActiveState string `json:"active_state"`
	SubState    string `json:"sub_state"`
	Description string `json:"description"`
}

// FetchServices fetches all systemd services from the remote VM.
func FetchServices(db *sql.DB, vmID string) ([]ServiceInfo, error) {
	session, err := sshclient.NewSession(db, vmID)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	// Parse plain output because --format=json might not be available on all systems.
	// Filter to --state=loaded to hide dead/not-found services.
	// awk splits by columns. The description is everything from column 5 onwards.
	cmd := `systemctl list-units --type=service --state=loaded --all --no-pager --no-legend | awk 'NF>=4 { desc=$5; for(i=6;i<=NF;i++) desc=desc " " $i; print $1 "---" $2 "---" $3 "---" $4 "---" desc }'`
	out, err := session.RunCmd(cmd)
	if err != nil {
		return nil, err
	}

	var services []ServiceInfo
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "---", 5)
		if len(parts) >= 4 {
			desc := ""
			if len(parts) == 5 {
				desc = strings.TrimSpace(parts[4])
			}
			services = append(services, ServiceInfo{
				Name:        strings.TrimSpace(parts[0]),
				LoadState:   strings.TrimSpace(parts[1]),
				ActiveState: strings.TrimSpace(parts[2]),
				SubState:    strings.TrimSpace(parts[3]),
				Description: desc,
			})
		}
	}

	return services, nil
}
