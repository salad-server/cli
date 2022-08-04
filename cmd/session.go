package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"time"

	gotmux "github.com/jubnzv/go-tmux"
	"github.com/salad-server/cli/utils"
)

func sendCommand(name, cmd string) error {
	if err := exec.Command("bash", "-c", fmt.Sprintf(
		`tmux send-keys -t "%s" "%s" ENTER`,
		name, cmd,
	)).Run(); err != nil {
		return err
	}

	return nil
}

func exitGrace(name string) error {
	if err := exec.Command("bash", "-c", fmt.Sprintf(`tmux send-keys -t "%s" C-c`, name)).Run(); err != nil {
		return err
	}

	return nil
}

func CreateSession(attach bool) error {
	server := new(gotmux.Server)
	exists, err := server.HasSession(utils.Config.SessionName)

	if err != nil {
		utils.Log.Error("[CreateSession] could not create session, ", err)
		return err
	}

	if exists {
		utils.Log.Fatalf("[CreateSession] %s already exists! Please kill it!", utils.Config.SessionName)
	}

	session := gotmux.Session{Name: utils.Config.SessionName}

	for id, s := range utils.Config.Sessions {
		session.AddWindow(gotmux.Window{
			Name:           s[0],
			Id:             id,
			StartDirectory: s[1],
		})
	}

	server.AddSession(session)

	sessions := []*gotmux.Session{}
	sessions = append(sessions, &session)
	conf := gotmux.Configuration{
		Server:        server,
		Sessions:      sessions,
		ActiveSession: nil,
	}

	if err = conf.Apply(); err != nil {
		utils.Log.Error("[CreateSession] could not apply session, ", err)
		return err
	}

	for _, cmd := range utils.Config.Sessions {
		if err := sendCommand(cmd[0], cmd[2]); err != nil {
			utils.Log.Error("[CreateSession] could not execute, ", err)
		}
	}

	if attach {
		if err = session.AttachSession(); err != nil {
			utils.Log.Error("[CreateSession] could not attach session, ", err)
			return err
		}
	} else {
		fmt.Println("Session created! Connect with:")
		fmt.Println("tmux a -t", utils.Config.SessionName)
	}

	return nil
}

func KillSessionSafe() error {
	server := new(gotmux.Server)
	exists, err := server.HasSession(utils.Config.SessionName)

	if err != nil {
		utils.Log.Error("[KillSessionSafe] could not create session, ", err)
		return err
	}

	if !exists {
		utils.Log.Fatalf("[KillSessionSafe] %s doesn't exist", utils.Config.SessionName)
		return errors.New("session not running")
	}

	for _, s := range utils.Config.Sessions {
		if err := exitGrace(s[0]); err != nil {
			utils.Log.Error("[KillSessionSafe] could not send ctrl+c for ", s[0], err)
			continue
		}

		if err := sendCommand(s[0], "exit"); err != nil {
			utils.Log.Error("[KillSessionSafe] could not send exit for ", s[0], err)
			continue
		}

		utils.Log.Info("sent commands | ", s[0])
	}

	return nil
}

func RestartSession(attach bool) error {
	server := new(gotmux.Server)
	exists, err := server.HasSession(utils.Config.SessionName)

	if err != nil {
		utils.Log.Error("[RestartSession] could not create session, ", err)
		return err
	}

	if !exists {
		return CreateSession(attach)
	}

	if err := KillSessionSafe(); err != nil {
		return err
	}

	for exists {
		exists, _ = server.HasSession(utils.Config.SessionName)
		utils.Log.Info("attempting to kill session, waiting 3s...")
		time.Sleep(3 * time.Second)
	}

	return CreateSession(attach)
}
