package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/salad-server/cli/structs"
	"github.com/salad-server/cli/utils"
	"golang.org/x/exp/slices"
)

// status on osu
var osu = map[string]string{
	// "-1": "notsubmitted",
	"0": "pending",
	"1": "ranked",
	"2": "approved",
	"3": "qualified",
	"4": "loved",
}

// bancho.py status
var mp = map[string]string{
	// "-1": "notsubmitted",
	"pending": "0",
	// "updateavailable": "1",
	"ranked":    "2",
	"approved":  "3",
	"qualified": "4",
	"loved":     "5",
}

func getBanchoStatus(id int) structs.BanchoBeatmap {
	var data []structs.BanchoBeatmap
	api := fmt.Sprintf("https://old.ppy.sh/api/get_beatmaps?k=%s&b=%d", utils.Config.API, id)
	req, err := http.Get(api)

	if err != nil {
		utils.Log.Fatal("[getBanchoStatus] could not get ", id, err)
	}

	defer req.Body.Close()

	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		utils.Log.Fatal("[getBanchoStatus] could not unmarshal ", id, err)
	}

	return data[0]
}

func updateDups(id int) error {
	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	_, err := utils.Database.ExecContext(c, "UPDATE maps SET status = 0 WHERE id = ?", id)

	defer cancel()

	if err != nil {
		utils.Log.Error("[updateDups] could not execute ", id, err)
		return err
	}

	return nil
}

func updateBeatmap(id int, status, md5 string) error {
	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	_, err := utils.Database.ExecContext(c, "UPDATE maps SET status = ? WHERE md5 = ? AND id = ?", status, md5, id)

	defer cancel()

	if err != nil {
		utils.Log.Error("[updateStatus] could not execute ", id, err)
		return err
	}

	return nil
}

func updateDatabase(bmap structs.DBeatmap) {
	banchoStatus := getBanchoStatus(bmap.ID)
	saladStatus := bmap.Status
	mpStr := map[string]string{
		"0": "pending",
		"2": "ranked",
		"3": "approved",
		"4": "qualified",
		"5": "loved",
	}

	if !slices.Contains([]string{"0", "1", "2", "3", "4"}, banchoStatus.RankStatus) {
		utils.Log.Warnf("osu api returned an invalid status (%s), ignoring...", banchoStatus.RankStatus)
		return
	}

	if mp[osu[banchoStatus.RankStatus]] == saladStatus {
		utils.Log.Warnf("%s [%s] matches status on bancho! ignoring...", bmap.Title, bmap.Version)
		return
	}

	utils.Log.Infof(
		"<%d> | Bancho: %s (%s) | Live: %s (%s) | updated!",
		bmap.ID, osu[banchoStatus.RankStatus],
		mp[osu[banchoStatus.RankStatus]], mpStr[saladStatus], saladStatus,
	)

	updateDups(bmap.ID) // sometimes qualified maps don't get removed...
	updateBeatmap(bmap.ID, mp[osu[banchoStatus.RankStatus]], bmap.MD5)
}

func UpdateSet(id int) error {
	var maps []structs.DBeatmap
	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	rows, err := utils.Database.QueryContext(c, `
		SELECT id, title, status, version, md5 FROM maps
		WHERE set_id = ?
	`, id)

	utils.Log.Trace("Updating set: ", id)
	defer cancel()

	if err != nil {
		utils.Log.Error("[UpdateSet] could not query ", id, err)
		return err
	}

	for rows.Next() {
		var bmap structs.DBeatmap

		if err := rows.Scan(
			&bmap.ID,
			&bmap.Title,
			&bmap.Status,
			&bmap.Version,
			&bmap.MD5,
		); err != nil {
			utils.Log.Error("[UpdateSet] could not scan ", bmap, err)
			return err
		}

		maps = append(maps, bmap)
	}

	for bmap := range maps {
		if !slices.Contains([]string{"0", "2", "3", "4", "5"}, maps[bmap].Status) {
			utils.Log.Warnf("%s contains an invalid status, ignoring...", maps[bmap].Status)
			continue
		}

		updateDatabase(maps[bmap])
	}

	return nil
}

func UpdateSetStatus(status string) error {
	var maps []structs.DBeatmap
	dbid := mp[status]
	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	rows, err := utils.Database.QueryContext(c, `
		SELECT id, title, status, version, md5 FROM maps
		WHERE status = ?
	`, dbid)

	utils.Log.Tracef("Updating maps with status %s (%s)", status, dbid)
	defer cancel()

	if err != nil {
		utils.Log.Error("[UpdateSetStatus] could not query ", dbid, err)
		return err
	}

	for rows.Next() {
		var bmap structs.DBeatmap

		if err := rows.Scan(
			&bmap.ID,
			&bmap.Title,
			&bmap.Status,
			&bmap.Version,
			&bmap.MD5,
		); err != nil {
			utils.Log.Error("[UpdateSet] could not scan ", bmap, err)
			return err
		}

		maps = append(maps, bmap)
	}

	for bmap := range maps {
		if !slices.Contains([]string{"0", "2", "3", "4", "5"}, maps[bmap].Status) {
			utils.Log.Warnf("%s contains an invalid status, ignoring...", maps[bmap].Status)
			continue
		}

		updateDatabase(maps[bmap])
	}

	return nil
}
