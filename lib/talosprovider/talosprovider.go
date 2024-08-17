// package talosprovider contains logic for installing talos on a server, which can vary by server provider
// The end result of installing is talos running in maintenance mode, ready to be added to a cluster.
package talosprovider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Server struct {
	Id                    string
	Ip                    string
	Username              string
	SshKeyPrivateBase64   string
	SshKeyPrivatePassword string
	SshKeyFingerprint     string
}

type InstallOptions struct {
	TalosVersion string
	Arch         string
}

type InstallOption func(*InstallOptions)

func WithTalosVersion(version string) InstallOption {
	return func(opts *InstallOptions) {
		opts.TalosVersion = version
	}
}

func WithArch(arch string) InstallOption {
	return func(opts *InstallOptions) {
		opts.Arch = arch
	}
}

func (o *InstallOptions) Validate() error {
	var errs []error
	if o.TalosVersion == "" {
		errs = append(errs, errors.New("talos version is required"))
	} else if !strings.HasPrefix(o.TalosVersion, "1.") || strings.Count(o.TalosVersion, ".") != 2 {
		errs = append(errs, errors.New("talos version must start with '1.' and have exactly two '.'s"))
	}
	if o.Arch == "" {
		errs = append(errs, errors.New("arch is required"))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

type TalosProvider interface {
	Install(ctx context.Context, server Server, opts ...InstallOption) error
}

type SshClient struct {
	Config *ssh.ClientConfig
	Server string
}

func NewSshClient(user string, host string, port int, privateKeyBase64 string, privateKeyPassword string) (*SshClient, error) {
	pemBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("decoding private key data failed %v", err)
	}
	// create signer
	signer, err := signerFromPem(pemBytes, []byte(privateKeyPassword))
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// use OpenSSH's known_hosts file if you care about host validation
			return nil
		},
		BannerCallback: func(message string) error {
			fmt.Println("DEBUG BANNER", message)
			return nil
		},
	}

	client := &SshClient{
		Config: config,
		Server: fmt.Sprintf("%v:%v", host, port),
	}

	return client, nil
}

// Opens a new SSH connection and runs the specified command
// Returns the combined output of stdout and stderr
func (s *SshClient) RunCommand(cmd string) (string, error) {
	// open connection
	conn, err := ssh.Dial("tcp", s.Server, s.Config)
	if err != nil {
		return "", fmt.Errorf("dial to %v failed %v", s.Server, err)
	}
	defer conn.Close()

	// open session
	session, err := conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("create session for %v failed %v", s.Server, err)
	}
	defer session.Close()
	// run command and capture stdout/stderr
	output, err := session.CombinedOutput(cmd)

	return string(output), err
}

func signerFromPem(pemBytes []byte, password []byte) (ssh.Signer, error) {
	pemBlock, _ := pem.Decode(pemBytes)
	if pemBlock == nil {
		return nil, errors.New("pem decode failed, no key found")
	}
	// generate signer instance from plain key
	signer, err := ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(password))
	if err != nil {
		return nil, fmt.Errorf("parsing plain private key failed %v", err)
	}
	return signer, nil
}

type diskInfo struct {
	Name   string
	SizeGB float64
}

func createAndRunCommand(server Server, command string) (string, error) {
	sshClient, err := NewSshClient(server.Username, server.Ip, 22, server.SshKeyPrivateBase64, server.SshKeyPrivatePassword)
	if err != nil {
		return "", fmt.Errorf("failed to create SSH client: %v", err)
	}

	output, err := sshClient.RunCommand(command)
	if err != nil {
		return "", fmt.Errorf("failed to run command '%s': %v, output: %s", command, err, output)
	}

	return output, nil
}

func getDiskInfo(server Server) ([]diskInfo, error) {
	output, err := createAndRunCommand(server, "lsblk -b -o NAME,SIZE,TYPE --noheadings --json")
	if err != nil {
		return nil, err
	}

	var result struct {
		BlockDevices []struct {
			Name string `json:"name"`
			Size int64  `json:"size"`
			Type string `json:"type"`
		} `json:"blockdevices"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("failed to parse lsblk output: %v", err)
	}

	var disks []diskInfo
	for _, device := range result.BlockDevices {
		if device.Type == "disk" {
			disks = append(disks, diskInfo{
				Name:   device.Name,
				SizeGB: float64(device.Size) / (1024 * 1024 * 1024),
			})
		}
	}

	return disks, nil
}

func SetupAndWipeFilesystem(server Server) error {
	disks, err := getDiskInfo(server)
	if err != nil {
		return err
	}

	for _, disk := range disks {
		commands := []string{
			fmt.Sprintf("parted --script -a optimal /dev/%s mklabel gpt", disk.Name),
			fmt.Sprintf("parted --script -a optimal /dev/%s mkpart primary ext4 0%% 100%%", disk.Name),
			fmt.Sprintf("mkfs.ext4 /dev/%sp1", disk.Name),
			fmt.Sprintf("sfdisk --delete /dev/%s", disk.Name),
			fmt.Sprintf("wipefs -a -f /dev/%s", disk.Name),
		}

		for _, cmd := range commands {
			if _, err := createAndRunCommand(server, cmd); err != nil {
				return fmt.Errorf("failed to run command '%s' on disk '%s': %v", cmd, disk.Name, err)
			}
		}
	}

	return nil
}

func StopRaids(server Server) error {
	output, err := createAndRunCommand(server, "mdadm --detail --scan")
	if err != nil {
		return err
	}

	var raids []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ARRAY") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				raids = append(raids, fields[1])
			}
		}
	}

	for _, raid := range raids {
		_, err := createAndRunCommand(server, fmt.Sprintf("mdadm --stop %s", raid))
		if err != nil {
			return fmt.Errorf("failed to stop RAID %s: %v", raid, err)
		}
	}

	return nil
}

func DownloadAndInstallTalos(server Server, version, flavor, arch string) error {
	disks, err := getDiskInfo(server)
	if err != nil {
		return err
	}

	if len(disks) == 0 {
		return errors.New("no disks found")
	}

	// Sort disks lexicographically by name
	sort.Slice(disks, func(i, j int) bool {
		return disks[i].Name < disks[j].Name
	})

	firstDisk := disks[0].Name
	url := fmt.Sprintf("https://github.com/siderolabs/talos/releases/download/v%s/%s-%s.raw.xz", version, flavor, arch)

	// Verify the release exists
	resp, err := http.Head(url)
	if err != nil {
		return fmt.Errorf("failed to verify Talos release: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("talos release not found at URL: %s", url)
	}

	commands := []string{
		fmt.Sprintf("wget %s -O /tmp/talos.xz", url),
		"xz -d -c /tmp/talos.xz | dd of=/dev/" + firstDisk,
		"sync",
	}

	for _, cmd := range commands {
		if _, err := createAndRunCommand(server, cmd); err != nil {
			return fmt.Errorf("failed to run command '%s' on disk '%s': %v", cmd, firstDisk, err)
		}
	}

	return nil
}
