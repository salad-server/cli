package cmd

import (
	"context"
	"time"

	"github.com/salad-server/cli/structs"
	"github.com/salad-server/cli/utils"
)

// Get score info
func getScoreInfo(id int) (structs.Score, error) {
	var info structs.Score
	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	score := utils.Database.QueryRowContext(c, `
		SELECT 
			s.mode, s.userid, s.map_md5,
			u.name
		FROM scores s
		JOIN users u ON s.userid = u.id
		WHERE s.id = ?	
	`, id)

	defer cancel()

	if err := score.Scan(
		&info.Mode,
		&info.UserID,
		&info.MD5,
		&info.Username,
	); err != nil {
		return info, err
	}

	return info, nil
}

// Set all scores to not pb
func setNotPB(uid, mode int, md5 string) error {
	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	_, err := utils.Database.ExecContext(c, `
		UPDATE scores SET status = 1
		WHERE userid = ? AND
			mode = ?     AND
			status != 0  AND
			map_md5 = ?
	`, uid, mode, md5)

	defer cancel()

	if err != nil {
		utils.Log.Error("[setNotPB] could not execute ", uid, err)
		return err
	}

	return nil
}

// Mark score <id> as pb
func setPB(id int) error {
	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	_, err := utils.Database.ExecContext(c, "UPDATE scores SET status = 2 WHERE id = ?", id)

	defer cancel()

	if err != nil {
		utils.Log.Error("[setPB] could not execute ", id, err)
		return err
	}

	return nil
}

// Set <id> as a personal best
func PersonalBest(id int) error {
	scoreInfo, err := getScoreInfo(id)

	// Should check if current score is personal best
	// but may serve as a way to remove multiple pbs.
	if err != nil {
		utils.Log.Error("[PersonalBest] could not get score ", id)
		return err
	}

	utils.Log.Infof(
		"%s | Settings %s's score <%d> as a personal best...",
		scoreInfo.MD5, scoreInfo.Username, id,
	)

	if err := setNotPB(scoreInfo.UserID, scoreInfo.Mode, scoreInfo.MD5); err != nil {
		return err
	}

	if err := setPB(id); err != nil {
		return err
	}

	return nil
}
