package ssh

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"time"

	gossh "golang.org/x/crypto/ssh"
	"backend/internal/crypto"
	"backend/internal/models"
)

const (
	dialTimeout = 10 * time.Second
	keepAlive   = 30 * time.Second
)

// SSHSession defines the interface for an SSH session to enable mocking in tests.
type SSHSession interface {
	RunCmd(cmd string) ([]byte, error)
	Run(cmd string) error
	RequestPty(term string, h, w int, modes gossh.TerminalModes) error
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.Reader, error)
	Shell() error
	Start(cmd string) error
	Wait() error
	WindowChange(rows, cols int) error
	Close()
}

// Session wraps an active SSH connection and its PTY session.
type Session struct {
	client  *gossh.Client
	session *gossh.Session
}

// NewSessionFunc is the type for the NewSession function to allow mocking.
type NewSessionFunc func(db *sql.DB, vmID string, loginAs ...string) (SSHSession, error)

// NewSession is a variable that defaults to the real SSH session creator.
var NewSession NewSessionFunc = func(db *sql.DB, vmID string, loginAs ...string) (SSHSession, error) {
	// 1. Fetch VM from DB (includes encrypted credential)
	vm, err := models.GetVMByID(db, vmID)
	if err != nil {
		return nil, fmt.Errorf("vm not found: %w", err)
	}

	aesKey := crypto.MustGetAESKey()

	// 2. Build SSH auth method — key vault takes priority over inline credential
	var authMethod gossh.AuthMethod

	if vm.KeyID != "" {
		// Path A: key vault — fetch and decrypt the stored private key
		authMethod, err = keyVaultAuth(db, vm.KeyID, aesKey)
		if err != nil {
			return nil, fmt.Errorf("key vault auth: %w", err)
		}
	} else {
		// Path B: inline credential (password or inline PEM key)
		plaintextCred, err := decryptCredential(vm.Credential, aesKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt credential: %w", err)
		}

		switch vm.AuthType {
		case "key":
			signer, err := gossh.ParsePrivateKey(plaintextCred)
			if err != nil {
				return nil, fmt.Errorf("failed to parse private key: %w", err)
			}
			authMethod = gossh.PublicKeys(signer)
		default: // "password"
			authMethod = gossh.Password(string(plaintextCred))
		}
	}

	// 3. Dial
	username := vm.ManagementUsername
	if len(loginAs) > 0 && loginAs[0] != "" {
		username = loginAs[0]
	}

	config := &gossh.ClientConfig{
		User:            username,
		Auth:            []gossh.AuthMethod{authMethod},
		// TODO: replace with known_hosts verification in production
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		Timeout:         dialTimeout,
	}

	host := vm.Host
	if _, _, err := net.SplitHostPort(host); err != nil {
		host = host + ":22"
	}

	conn, err := net.DialTimeout("tcp", host, dialTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", host, err)
	}

	sshConn, chans, reqs, err := gossh.NewClientConn(conn, host, config)
	if err != nil {
		return nil, fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	client := gossh.NewClient(sshConn, chans, reqs)

	// 4. Open an interactive session (PTY + shell start handled by caller)
	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}

	return &Session{
		client:  client,
		session: session,
	}, nil
}


// keyVaultAuth fetches and decrypts a key from the ssh_keys table and returns
// a gossh.AuthMethod using that key as a signer.
func keyVaultAuth(db *sql.DB, keyID string, aesKey []byte) (gossh.AuthMethod, error) {
	key, err := models.GetSSHKeyByID(db, keyID)
	if err != nil {
		return nil, fmt.Errorf("key not found in vault (id=%s): %w", keyID, err)
	}

	plaintext, err := decryptCredential(key.PrivateKey, aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt vault key: %w", err)
	}

	signer, err := gossh.ParsePrivateKey(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vault private key: %w", err)
	}

	return gossh.PublicKeys(signer), nil
}

// decryptCredential base64-decodes and AES-decrypts an encrypted credential string.
func decryptCredential(encBase64 string, aesKey []byte) ([]byte, error) {
	encBytes, err := base64.StdEncoding.DecodeString(encBase64)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}
	return crypto.Decrypt(encBytes, aesKey)
}

// RequestPty allocates a pseudo-terminal on the remote server.
func (s *Session) RequestPty(term string, h, w int, modes gossh.TerminalModes) error {
	return s.session.RequestPty(term, h, w, modes)
}

// Shell starts an interactive login shell on the remote server.
func (s *Session) Shell() error {
	return s.session.Shell()
}

// RunCmd executes a single command on the remote server and returns its output.
// Note: An SSH session can only run one command. The session is consumed.
func (s *Session) RunCmd(cmd string) ([]byte, error) {
	return s.session.CombinedOutput(cmd)
}

// Run starts a remote command and waits for it to finish.
func (s *Session) Run(cmd string) error {
	return s.session.Run(cmd)
}

// Start starts a remote command without waiting for it to finish.
func (s *Session) Start(cmd string) error {
	return s.session.Start(cmd)
}

// Wait waits for the remote command to exit.
func (s *Session) Wait() error {
	return s.session.Wait()
}

// StdinPipe returns a write-closer for the SSH session's stdin.
func (s *Session) StdinPipe() (io.WriteCloser, error) {
	return s.session.StdinPipe()
}

// StdoutPipe returns a reader for the SSH session's combined stdout+stderr.
func (s *Session) StdoutPipe() (io.Reader, error) {
	return s.session.StdoutPipe()
}

// WindowChange notifies the remote end of a terminal resize event.
func (s *Session) WindowChange(rows, cols int) error {
	return s.session.WindowChange(rows, cols)
}

// Close terminates the SSH session and the underlying connection.
func (s *Session) Close() {
	if s.session != nil {
		s.session.Close()
	}
	if s.client != nil {
		s.client.Close()
	}
}

// RunSystemdCommand executes a systemctl command for a given service.
func RunSystemdCommand(db *sql.DB, vmID, service, action string) ([]byte, error) {
	validActions := map[string]bool{"start": true, "stop": true, "restart": true, "status": true, "enable": true, "disable": true}
	if !validActions[action] {
		return nil, fmt.Errorf("invalid action: %s", action)
	}

	// Basic sanitation for service name
	for _, c := range service {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.') {
			return nil, fmt.Errorf("invalid service name characters")
		}
	}

	session, err := NewSession(db, vmID)
	if err != nil {
		return nil, fmt.Errorf("ssh connection failed: %w", err)
	}
	defer session.Close()

	cmd := fmt.Sprintf("sudo systemctl %s %s", action, service)
	return session.RunCmd(cmd)
}

// RunPowerAction executes a reboot, shutdown, or sleep command on the remote server.
func RunPowerAction(db *sql.DB, vmID, action string) error {
	var cmd string
	switch action {
	case "reboot":
		cmd = "sudo /usr/sbin/reboot"
	case "shutdown":
		cmd = "sudo /usr/sbin/poweroff"
	case "sleep":
		cmd = "sudo /usr/bin/systemctl suspend"
	default:
		return fmt.Errorf("invalid power action: %s", action)
	}

	session, err := NewSession(db, vmID)
	if err != nil {
		return fmt.Errorf("ssh connection failed: %w", err)
	}
	defer session.Close()

	// Execute command. We expect a connection error (EOF or reset) as the server closes the SSH session.
	// Using session.Run directly to handle the expected connection drop.
	err = session.Run(cmd)
	if err != nil {
		// If the error is EOF or connection reset, we consider it a success for power actions
		// because the server is actively closing the connection to reboot/shutdown.
		errStr := err.Error()
		if errStr == "EOF" || 
		   errStr == "connection reset by peer" || 
		   errStr == "ssh: disconnect" ||
		   errStr == "waitid: no child processes" {
			return nil
		}
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}
