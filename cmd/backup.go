package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/mholt/archiver/v4"
	"github.com/salad-server/cli/utils"
)

// Dump SQL
func dumpSQL(filename string) error {
	dsn, err := mysql.ParseDSN(utils.Config.DSN)

	if err != nil {
		utils.Log.Error("[dumpSQL] could not parse dsn ", dsn)
		return err
	}

	if err := exec.Command("bash", "-c", fmt.Sprintf(
		"mysqldump --user='%s' --password='%s' %s > %s",
		dsn.User, dsn.Passwd, dsn.DBName, filename,
	)).Run(); err != nil {
		utils.Log.Error("[dumpSQL] could not dump SQL ", err)
		return err
	}

	return nil
}

func exportCompress(filenames map[string]string, output string) error {
	files, err := archiver.FilesFromDisk(nil, filenames)
	if err != nil {
		return err
	}

	out, err := os.Create("./data/" + output + ".tar.gz")
	if err != nil {
		return err
	}

	defer out.Close()

	format := archiver.CompressedArchive{
		Compression: archiver.Gz{},
		Archival:    archiver.Tar{},
	}

	if err = format.Archive(context.Background(), out, files); err != nil {
		return err
	}

	fmt.Printf("Success! Check data folder for %s.tar.gz\n", output)
	return nil
}

// Make and copy backup into ./data/<date>
func Backup(sql, replays, data bool) error {
	var wg sync.WaitGroup
	var filename = time.Now().Format("2006-01-02")

	archive := make(map[string]string)

	if sql {
		utils.Log.Info("Dumping SQL...")
		wg.Add(1)

		go (func() {
			outname, err := filepath.Abs("./data/" + filename + ".sql")

			defer wg.Done()
			if err != nil {
				utils.Log.Error("[Backup] could not resolve absolute path", err)
				return
			}

			archive[outname] = filename + ".sql"
			dumpSQL(outname)
		})()
	}

	if replays {
		utils.Log.Info("Copying replays...")
		archive[utils.Config.Data.Replays] = "osr"
	}

	if data {
		utils.Log.Info("Copying data (ss, assets, avatars)...")
		archive[utils.Config.Data.Screenshots] = "ss"
		archive[utils.Config.Data.Avatars] = "avatars"
	}

	wg.Wait()
	compressed := exportCompress(archive, filename)

	if sql {
		if err := os.Remove("./data/" + filename + ".sql"); err != nil {
			utils.Log.Error("[Backup] could not remove SQL dump", err)
		}
	}

	return compressed
}
