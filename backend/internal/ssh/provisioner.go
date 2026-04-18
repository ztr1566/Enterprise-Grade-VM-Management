package ssh

import (
	"database/sql"
	_ "embed"
	"fmt"
	"backend/internal/audit"
	"backend/internal/models"
	"go.uber.org/zap"
)

//go:embed provisioning/setup_vm_sudoers.sh
var provisioningScript string

// ProvisionVM executes the provisioning script on the target VM to configure scoped sudoers rules.
// It uses the management_username to connect and provisions all users listed in vm_access.
func ProvisionVM(db *sql.DB, vmID string) error {
	// 1. Fetch VM record
	vm, err := models.GetVMByID(db, vmID)
	if err != nil {
		return fmt.Errorf("fetching vm: %w", err)
	}

	// 2. Fetch authorized users from vm_access
	usernames, err := models.GetUsersByVM(db, vmID)
	if err != nil {
		audit.Logger.Error("Failed to fetch VM users for provisioning", zap.String("vmID", vmID), zap.Error(err))
		// We'll proceed and let the fallback logic handle it
	}

	// Fallback to management_username if no specific users found
	if len(usernames) == 0 {
		usernames = []string{vm.ManagementUsername}
	}

	// 3. Initialize provisioning state
	models.UpdateProvisioningStatus(db, vmID, "pending", nil, nil)
	models.CreateAuditLog(db, "VM_PROVISIONING_START", fmt.Sprintf("Provisioning started for VM %s (%s)", vm.Name, vmID))

	// 4. Upload script via SSH to /tmp (using management credentials)
	session, err := NewSession(db, vmID)
	if err != nil {
		errMsg := err.Error()
		models.UpdateProvisioningStatus(db, vmID, "failed", &errMsg, nil)
		models.CreateAuditLog(db, "VM_PROVISIONING_FAILED", fmt.Sprintf("Provisioning failed for VM %s: %s", vm.Name, errMsg))
		return err
	}

	tempPath := "/tmp/setup_vm_sudoers.sh"
	uploadCmd := fmt.Sprintf("cat > %s << 'HEREDOC'\n%s\nHEREDOC", tempPath, provisioningScript)
	output, err := session.RunCmd(uploadCmd)
	if err != nil {
		session.Close()
		errMsg := fmt.Sprintf("upload failed: %v. Output: %s", err, string(output))
		models.UpdateProvisioningStatus(db, vmID, "failed", &errMsg, nil)
		models.CreateAuditLog(db, "VM_PROVISIONING_FAILED", fmt.Sprintf("Provisioning failed for VM %s: %s", vm.Name, errMsg))
		return err
	}
	session.Close()

	// 5. Execute script via SSH for each user
	var lastErr error
	var successCount int

	for _, username := range usernames {
		// Connect as management user to provision the target user
		execSession, err := NewSession(db, vmID)
		if err != nil {
			lastErr = fmt.Errorf("connection failed for user %s: %w", username, err)
			audit.Logger.Error("Multi-user provisioning: connection failed", zap.String("username", username), zap.Error(err))
			continue
		}

		execCmd := fmt.Sprintf("sudo bash %s %s", tempPath, username)
		output, err := execSession.RunCmd(execCmd)
		execSession.Close()

		if err != nil {
			lastErr = fmt.Errorf("execution failed for user %s: %v. Output: %s", username, err, string(output))
			models.CreateAuditLog(db, "VM_USER_PROVISIONING_FAILED", fmt.Sprintf("Provisioning user %s failed for VM %s: %v", username, vm.Name, lastErr))
			audit.Logger.Error("Multi-user provisioning: execution failed", zap.String("username", username), zap.Error(err))
			continue
		}
		successCount++
	}

	// Best-effort cleanup
	if cleanSession, err := NewSession(db, vmID); err == nil {
		cleanupCmd := fmt.Sprintf("rm -f %s", tempPath)
		_, _ = cleanSession.RunCmd(cleanupCmd)
		cleanSession.Close()
	}

	if successCount == 0 && len(usernames) > 0 {
		errMsg := lastErr.Error()
		models.UpdateProvisioningStatus(db, vmID, "failed", &errMsg, nil)
		return lastErr
	}

	// 6. Finalize status
	models.UpdateProvisioningStatus(db, vmID, "provisioned", nil, &vm.ManagementUsername)
	models.CreateAuditLog(db, "VM_PROVISIONED", fmt.Sprintf("Provisioning completed successfully for VM %s", vm.Name))

	return nil
}

// RevokeUserAccess removes the sudoers file for a specific user from the target VM.
func RevokeUserAccess(db *sql.DB, vmID, username string) error {
	session, err := NewSession(db, vmID)
	if err != nil {
		return fmt.Errorf("ssh connection failed: %w", err)
	}
	defer session.Close()

	// Delete the sudoers.d file
	revokeCmd := fmt.Sprintf("sudo rm -f /etc/sudoers.d/vm-platform-%s", username)
	output, err := session.RunCmd(revokeCmd)
	if err != nil {
		return fmt.Errorf("revoke command failed: %v. Output: %s", err, string(output))
	}

	audit.Logger.Info("User access revoked via SSH", zap.String("vmID", vmID), zap.String("username", username))
	return nil
}
