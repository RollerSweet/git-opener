package tmuxxer

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Tmuxxer struct{}

func IsTmuxSession() (bool, error) {
	output, err := exec.Command("/bin/bash", "-c", "echo $TMUX").
		CombinedOutput()

	if err != nil {
		return false, fmt.Errorf("Failed to check if tmux session %w", err)
	}

	return len(output) != 1, nil
}

func HasSession(name string) bool {
	_, err := exec.Command("tmux", "has-session", "-t", name).
		CombinedOutput()

	if err != nil {
		return false
	}

	return true
}

func GetWindowsAmount(session_name string) (int, error) {
	res := exec.Command("/bin/bash", "-c", fmt.Sprintf("tmux list-windows -t %s | wc -l", session_name))
	output, err := res.CombinedOutput()
	if err != nil {
		return 0, err
	}

	parsed_output := strings.TrimSpace(string(output))
	num, err := strconv.Atoi(parsed_output)
	if err != nil {
		return 0, err
	}

	return num, nil
}

func CreateSession(name string, detach bool) error {
	cmd_builder := []string{"tmux", "new-session"}

	if detach {
		cmd_builder = append(cmd_builder, "-d")
	}
	cmd_builder = append(cmd_builder, "-s")
	cmd_builder = append(cmd_builder, fmt.Sprintf("'%s'", name))

	cmd := exec.Command("/bin/bash", "-c", strings.Join(cmd_builder, " "))
	return cmd.Run()
}

func SendKeys(session_name string, window_number string, cmd string) error {
	cmd_res := exec.Command("tmux", "send-keys", "-t", fmt.Sprintf("%s:%s", session_name, window_number), fmt.Sprintf("%s", cmd), "C-m")
	return cmd_res.Run()
}

func ChangeSession(session_name string) error {
	is_session, err := IsTmuxSession()
	arg := "attach-session"

	if err != nil {
		return err
	}

	if is_session {
		arg = "switch-client"
	}

	return exec.Command("tmux", arg, "-t", session_name).Run()
}

func CreateWindow(session_name string, window_name string) error {
	cmd_builder := []string{"tmux", "new-window", "-t", session_name, "-n", window_name}
	exec.Command("/bin/bash", "-c", strings.Join(cmd_builder, " ")).Run()
	last_window, _ := GetWindowsAmount(session_name)
	cmd := exec.Command("tmux", "rename-window", "-t", fmt.Sprintf("%s:%d", session_name, last_window), window_name)
	return cmd.Run()
}

func SelectWindow(session_name string, window_number string) error {
	return exec.Command("tmux", "select-window", "-t", fmt.Sprintf("%s:%s", session_name, window_number)).Run()
}

// ListSessions returns a list of all active tmux sessions
func ListSessions() ([]string, error) {
	output, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name}").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list tmux sessions: %w", err)
	}

	sessions := strings.Split(strings.TrimSpace(string(output)), "\n")
	// Filter out empty strings
	var result []string
	for _, session := range sessions {
		if session != "" {
			result = append(result, session)
		}
	}

	return result, nil
}

// GetCurrentSession returns the name of the current tmux session
func GetCurrentSession() (string, error) {
	output, err := exec.Command("tmux", "display-message", "-p", "#{session_name}").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get current tmux session: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
