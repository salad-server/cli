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

// Status
var mp = map[string]string{
	// "not_submitted": "-1",
	// "update_available": "1",
	"pending":   "0",
	"ranked":    "2",
	"approved":  "3",
	"qualified": "4",
	"loved":     "5",
}

var mpStr = map[string]string{
	"0": "pending",
	"2": "ranked",
	"3": "approved",
	"4": "qualified",
	"5": "loved",
}

// Get beatmap's status on osu.ppy.sh
func getBanchoStatus(id int) structs.BanchoBeatmap {
	var data []structs.BanchoBeatmap
	api := fmt.Sprintf("https://old.ppy.sh/api/get_beatmaps?k=%s&b=%d", utils.Config.API, id)
	req, err := http.Get(api)

	if err != nil {
		utils.Log.Fatal("[getBanchoStatus] could not get ", id, err)
	}

	// TODO: Allow for multiple API keys
	defer req.Body.Close()

	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		utils.Log.Fatal("[getBanchoStatus] could not unmarshal ", id, err)
	}

	return data[0]
}

// Update all maps in DB to pending
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

// Update all maps (with proper md5) to <status>
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

// Convert osu.ppy.sh status to bancho.py status
// https://github.com/osuAkatsuki/bancho.py/blob/master/app/objects/beatmap.py#L102
func convertStatus(status string) string {
	switch status {
	// Pending
	case "-2", "-1", "-": // Graveyard, NotSubmitted, Pending
		return "0"

	// Ranked
	case "1":
		return "2"

	// Approved
	case "2":
		return "3"

	// Qualified
	case "3":
		return "4"

	// Loved
	case "4":
		return "5"

	// what??
	default:
		utils.Log.Fatal("Invalid status!", status)
	}

	return ""
}

// Update a beatmap with status from osu.ppy.sh
func updateDatabase(bmap structs.DBeatmap) {
	banchoStatus := getBanchoStatus(bmap.ID)
	newStatus := convertStatus(banchoStatus.RankStatus)
	saladStatus := bmap.Status

	if saladStatus == newStatus {
		utils.Log.Warnf("<%d> | Beatmap is already %s (%s)", bmap.ID, newStatus, mpStr[newStatus])
		return
	}

	utils.Log.Infof("<%d> | %s [%s] | %s -> %s | %s", bmap.ID, bmap.Title, bmap.Version, saladStatus, newStatus, mpStr[newStatus])
	updateDups(bmap.ID)
	updateBeatmap(bmap.ID, newStatus, bmap.MD5)
}

// Fetch every map in set, run updateDatabase() on every difficulty.
func UpdateSet(id int) error {
	var maps []structs.DBeatmap
	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	rows, err := utils.Database.QueryContext(c, `
		SELECT id, title, status, version, md5 FROM maps
		WHERE set_id = ?
	`, id)

	fmt.Println("Updating set: ", id)
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

// Fetch every map with <status>, run updateDatabase() on every beatmap.
func UpdateSetStatus(status string) error {
	var maps []structs.DBeatmap
	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	rows, err := utils.Database.QueryContext(c, `
		SELECT id, title, status, version, md5 FROM maps
		WHERE status = ?
	`, mp[status])

	fmt.Printf("Updating maps with status %s (%s)\n", status, mp[status])
	defer cancel()

	if err != nil {
		utils.Log.Error("[UpdateSetStatus] could not query ", mp[status], err)
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
