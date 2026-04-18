package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"backend/internal/api"
	"backend/internal/audit"
	"backend/internal/models"
	sshclient "backend/internal/ssh"
)

type SecurityHandler struct {
	DB *sql.DB
}

// ---- Response types ----

type FirewallStatus struct {
	Backend string         `json:"backend"` // "firewalld" | "ufw" | "NOT_INSTALLED"
	Active  bool           `json:"active"`
	Rules   []FirewallRule `json:"rules"`
}

type FirewallRule struct {
	Port     string `json:"port"`
	Protocol string `json:"protocol"`
	Action   string `json:"action"` // ACCEPT / DROP / REJECT
	Source   string `json:"source,omitempty"`
}

type SELinuxStatus struct {
	Status string `json:"status"` // "INSTALLED" | "NOT_INSTALLED"
	Mode   string `json:"mode"`   // enforcing | permissive | disabled
	Policy string `json:"policy"` // targeted | mls | …
}

type SecurityOverview struct {
	Firewall FirewallStatus `json:"firewall"`
	SELinux  SELinuxStatus  `json:"selinux"`
}

// ---- GET /api/vms/{id}/security ----

func (h *SecurityHandler) GetSecurityStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "Missing VM ID")
		return
	}

	overview := SecurityOverview{}

	// --- Firewall ---
	overview.Firewall = h.fetchFirewall(id)

	// --- SELinux ---
	overview.SELinux = h.fetchSELinux(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(overview)
}

// fetchFirewall tries firewalld first, then ufw.
// IMPORTANT: Each SSH session can only run ONE command (RunCmd consumes it).
// We use compound shell commands to do multi-step detection in a single call.
func (h *SecurityHandler) fetchFirewall(vmID string) FirewallStatus {
	status := FirewallStatus{Backend: "NOT_INSTALLED", Rules: []FirewallRule{}}

	// Single compound command: check which backend exists and whether it's active.
	// Output format: "BACKEND:STATE" e.g. "firewalld:active", "ufw:active", "none:none"
	session, err := sshclient.NewSession(h.DB, vmID)
	if err != nil {
		return status
	}
	defer session.Close()

	// This script checks firewalld first (via systemctl), then ufw, in one shot.
	detectCmd := `
if command -v firewall-cmd >/dev/null 2>&1; then
  STATE=$(systemctl is-active firewalld 2>/dev/null || echo "inactive")
  echo "firewalld:$STATE"
elif command -v ufw >/dev/null 2>&1; then
  UFW_OUT=$(sudo ufw status 2>/dev/null)
  if echo "$UFW_OUT" | grep -q "Status: active"; then
    echo "ufw:active"
  else
    echo "ufw:inactive"
  fi
else
  echo "none:none"
fi
`
	out, err := session.RunCmd(detectCmd)
	if err != nil {
		return status
	}

	result := strings.TrimSpace(string(out))
	parts := strings.SplitN(result, ":", 2)
	if len(parts) != 2 {
		return status
	}

	backend := parts[0]
	state := parts[1]

	switch backend {
	case "firewalld":
		status.Backend = "firewalld"
		status.Active = (state == "active")
		if status.Active {
			status.Rules = h.fetchFirewalldRules(vmID)
		}
	case "ufw":
		status.Backend = "ufw"
		status.Active = (state == "active")
		if status.Active {
			// Need to fetch UFW rules with a separate session
			s2, err := sshclient.NewSession(h.DB, vmID)
			if err == nil {
				defer s2.Close()
				ufwOut, err := s2.RunCmd("sudo ufw status 2>/dev/null")
				if err == nil {
					status.Rules = h.parseUFWRules(string(ufwOut))
				}
			}
		}
	}

	return status
}

func (h *SecurityHandler) fetchFirewalldRules(vmID string) []FirewallRule {
	session, err := sshclient.NewSession(h.DB, vmID)
	if err != nil {
		return []FirewallRule{}
	}
	defer session.Close()

	out, err := session.RunCmd("sudo firewall-cmd --list-ports 2>/dev/null")
	if err != nil {
		return []FirewallRule{}
	}

	var rules []FirewallRule
	ports := strings.Fields(strings.TrimSpace(string(out)))
	for _, p := range ports {
		parts := strings.SplitN(p, "/", 2)
		if len(parts) == 2 {
			rules = append(rules, FirewallRule{
				Port:     parts[0],
				Protocol: parts[1],
				Action:   "ACCEPT",
			})
		}
	}
	if rules == nil {
		rules = []FirewallRule{}
	}
	return rules
}

func (h *SecurityHandler) parseUFWRules(output string) []FirewallRule {
	var rules []FirewallRule
	lines := strings.Split(output, "\n")
	// UFW format: "80/tcp                     ALLOW       Anywhere"
	re := regexp.MustCompile(`^(\d+(?:/\w+)?)\s+(ALLOW|DENY|REJECT|LIMIT)\s+(.*)`)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		m := re.FindStringSubmatch(line)
		if m != nil {
			portProto := strings.SplitN(m[1], "/", 2)
			port := portProto[0]
			proto := "tcp"
			if len(portProto) == 2 {
				proto = portProto[1]
			}
			rules = append(rules, FirewallRule{
				Port:     port,
				Protocol: proto,
				Action:   m[2],
				Source:   strings.TrimSpace(m[3]),
			})
		}
	}
	if rules == nil {
		rules = []FirewallRule{}
	}
	return rules
}

// fetchSELinux detects SELinux status using a single compound SSH command.
// IMPORTANT: SSH sessions are single-use; RunCmd consumes the session.
func (h *SecurityHandler) fetchSELinux(vmID string) SELinuxStatus {
	status := SELinuxStatus{Status: "NOT_INSTALLED", Mode: "disabled", Policy: ""}

	session, err := sshclient.NewSession(h.DB, vmID)
	if err != nil {
		return status
	}
	defer session.Close()

	// Single compound command: find sestatus/getenforce (including /usr/sbin),
	// then run sestatus for full detail or fall back to getenforce.
	detectCmd := `
SESTATUS_CMD=""
GETENFORCE_CMD=""
if command -v sestatus >/dev/null 2>&1; then
  SESTATUS_CMD="sestatus"
elif test -x /usr/sbin/sestatus; then
  SESTATUS_CMD="/usr/sbin/sestatus"
fi
if command -v getenforce >/dev/null 2>&1; then
  GETENFORCE_CMD="getenforce"
elif test -x /usr/sbin/getenforce; then
  GETENFORCE_CMD="/usr/sbin/getenforce"
fi

if [ -z "$SESTATUS_CMD" ] && [ -z "$GETENFORCE_CMD" ]; then
  echo "NOT_INSTALLED"
  exit 0
fi

echo "INSTALLED"

if [ -n "$SESTATUS_CMD" ]; then
  sudo $SESTATUS_CMD 2>/dev/null
else
  MODE=$(sudo $GETENFORCE_CMD 2>/dev/null)
  echo "Current mode:                 $MODE"
fi
`
	out, err := session.RunCmd(detectCmd)
	if err != nil {
		return status
	}

	output := strings.TrimSpace(string(out))
	lines := strings.Split(output, "\n")

	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "NOT_INSTALLED" {
		return status
	}

	status.Status = "INSTALLED"

	// Parse sestatus-style output (lines after the first "INSTALLED" marker)
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Current mode:") {
			status.Mode = strings.TrimSpace(strings.TrimPrefix(line, "Current mode:"))
		}
		if strings.HasPrefix(line, "Loaded policy name:") {
			status.Policy = strings.TrimSpace(strings.TrimPrefix(line, "Loaded policy name:"))
		}
		if strings.HasPrefix(line, "SELinux status:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "SELinux status:"))
			if val == "disabled" {
				status.Mode = "disabled"
			}
		}
	}

	return status
}

// ---- POST /api/vms/{id}/firewall/rules ----

type FirewallRuleRequest struct {
	Action   string `json:"action"`   // "add" | "remove"
	Port     string `json:"port"`     // e.g. "8080"
	Protocol string `json:"protocol"` // "tcp" | "udp"
}

func (h *SecurityHandler) ManageFirewallRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "Missing VM ID")
		return
	}

	var req FirewallRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate port (1-65535, numeric only)
	portValid := regexp.MustCompile(`^\d{1,5}$`)
	if !portValid.MatchString(req.Port) {
		api.WriteError(w, http.StatusBadRequest, "Invalid port number")
		return
	}
	if req.Protocol != "tcp" && req.Protocol != "udp" {
		api.WriteError(w, http.StatusBadRequest, "Protocol must be 'tcp' or 'udp'")
		return
	}
	if req.Action != "add" && req.Action != "remove" {
		api.WriteError(w, http.StatusBadRequest, "Action must be 'add' or 'remove'")
		return
	}

	userID, _ := r.Context().Value("user_id").(string)

	// Detect which backend is in use
	fw := h.fetchFirewall(id)

	var cmd string
	switch fw.Backend {
	case "firewalld":
		flag := "--add-port"
		if req.Action == "remove" {
			flag = "--remove-port"
		}
		cmd = fmt.Sprintf("sudo firewall-cmd %s=%s/%s --permanent && sudo firewall-cmd --reload", flag, req.Port, req.Protocol)
	case "ufw":
		ufwAction := "allow"
		if req.Action == "remove" {
			ufwAction = "delete allow"
		}
		cmd = fmt.Sprintf("sudo ufw %s %s/%s", ufwAction, req.Port, req.Protocol)
	default:
		api.WriteError(w, http.StatusBadRequest, "No supported firewall detected on this VM")
		return
	}

	session, err := sshclient.NewSession(h.DB, id)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "SSH connection failed: "+err.Error())
		return
	}
	defer session.Close()

	out, cmdErr := session.RunCmd(cmd)

	// Audit
	details := fmt.Sprintf("user=%s vm=%s firewall_action=%s port=%s/%s backend=%s success=%v",
		userID, id, req.Action, req.Port, req.Protocol, fw.Backend, cmdErr == nil)
	if auditErr := models.CreateAuditLog(h.DB, "FIREWALL_RULE_MODIFIED", details); auditErr != nil {
		audit.Logger.Error("Failed to write audit log for firewall rule", zap.Error(auditErr))
	}

	if cmdErr != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"output": string(out),
			"error":  cmdErr.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"output": string(out),
	})
}

// ---- POST /api/vms/{id}/selinux ----

type SELinuxRequest struct {
	Mode string `json:"mode"` // "enforcing" | "permissive"
}

func (h *SecurityHandler) SetSELinuxMode(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "Missing VM ID")
		return
	}

	var req SELinuxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	modeMap := map[string]string{"enforcing": "1", "permissive": "0"}
	modeVal, ok := modeMap[req.Mode]
	if !ok {
		api.WriteError(w, http.StatusBadRequest, "Mode must be 'enforcing' or 'permissive'")
		return
	}

	userID, _ := r.Context().Value("user_id").(string)

	session, err := sshclient.NewSession(h.DB, id)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "SSH connection failed: "+err.Error())
		return
	}
	defer session.Close()

	cmd := fmt.Sprintf("sudo setenforce %s", modeVal)
	out, cmdErr := session.RunCmd(cmd)

	// Audit
	details := fmt.Sprintf("user=%s vm=%s selinux_mode=%s success=%v", userID, id, req.Mode, cmdErr == nil)
	if auditErr := models.CreateAuditLog(h.DB, "SELINUX_MODE_CHANGED", details); auditErr != nil {
		audit.Logger.Error("Failed to write audit log for SELinux change", zap.Error(auditErr))
	}

	if cmdErr != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"output": string(out),
			"error":  cmdErr.Error(),
		})
		return
	}

	audit.Logger.Info("SELinux mode changed",
		zap.String("vmID", id), zap.String("mode", req.Mode), zap.String("userID", userID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"mode":   req.Mode,
	})
}

// ---- POST /api/vms/{id}/security/install ----

type SecurityInstallRequest struct {
	Component string `json:"component"` // "firewall" | "selinux"
}

func (h *SecurityHandler) InstallSecurityComponent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "Missing VM ID")
		return
	}

	var req SecurityInstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	var cmd string
	switch req.Component {
	case "firewall":
		// stick to yum for now as per enterprise spec, but we can detect apt too
		cmd = "sudo yum install -y firewalld && sudo systemctl enable --now firewalld"
	case "selinux":
		cmd = "sudo yum install -y policycoreutils policycoreutils-python-utils selinux-policy selinux-policy-targeted"
	default:
		api.WriteError(w, http.StatusBadRequest, "Invalid component")
		return
	}

	userID, _ := r.Context().Value("user_id").(string)

	session, err := sshclient.NewSession(h.DB, id)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "SSH connection failed: "+err.Error())
		return
	}
	defer session.Close()

	out, cmdErr := session.RunCmd(cmd)

	// Audit
	details := fmt.Sprintf("user=%s vm=%s install_component=%s success=%v", userID, id, req.Component, cmdErr == nil)
	if auditErr := models.CreateAuditLog(h.DB, "SECURITY_COMPONENT_INSTALLED", details); auditErr != nil {
		audit.Logger.Error("Failed to write audit log for security install", zap.Error(auditErr))
	}

	if cmdErr != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"output": string(out),
			"error":  cmdErr.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"output": string(out),
	})
}

