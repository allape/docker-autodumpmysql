package main

import (
	"compress/gzip"
	"errors"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"slices"
	"syscall"
	"time"

	"github.com/allape/goenv"
	"github.com/allape/gogger"
	"github.com/robfig/cron/v3"
)

var l = gogger.New("autodump")

var (
	MySQLRootPassword = goenv.Getenv("MYSQL_ROOT_PASSWORD", "")
	MySQLUser         = goenv.Getenv("MYSQL_USER", "root")
	MySQLPassword     = goenv.Getenv("MYSQL_PASSWORD", "")
	MySQLDatabase     = goenv.Getenv("MYSQL_DATABASE", "")
	MySQLHomeDir      = goenv.Getenv("MYSQL_HOME_DIR", "/var/lib/mysql")

	MySQLAutodumpMySQLDumpBin = goenv.Getenv("MYSQL_AUTODUMP_MYSQL_DUMP_BIN", "/usr/bin/mysqldump")
	MySQLAutodumpOutputDir    = goenv.Getenv("MYSQL_AUTODUMP_OUTPUT_DIR", "autodump")
	MySQLAutodumpCron         = goenv.Getenv("MYSQL_AUTODUMP_CRON", "0 4 * * *") // 4:00 am of every morning, https://pkg.go.dev/github.com/robfig/cron?utm_source=godoc#hdr-CRON_Expression_Format
)

func NewDumper() func() {
	password := MySQLPassword
	if password == "" && MySQLUser == "root" {
		password = MySQLRootPassword
	}

	args := []string{
		"-u" + MySQLUser,
		"-p" + password,
	}

	if MySQLDatabase == "" {
		args = append(args, "--all-databases")
	} else {
		args = append(args, "--databases", MySQLDatabase)
	}
	return func() {
		l.Info().Printf("running autodump")

		//l.Info().Printf("%s %v", MySQLAutodumpMySQLDumpBin, args)
		cmd := exec.Command(MySQLAutodumpMySQLDumpBin, args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			l.Error().Printf("failed to run mysqldump: %s\n%s", err, out)
			return
		}

		now := time.Now().Format("20060102150405")

		fullPath := path.Join(MySQLHomeDir, MySQLAutodumpOutputDir, now+".sql.gz")
		dirPath := path.Dir(fullPath)

		stat, err := os.Stat(dirPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				err = os.MkdirAll(dirPath, 0755)
				if err != nil {
					l.Error().Printf("failed to create directory %s: %s", dirPath, err)
					return
				}
			} else {
				l.Error().Printf("failed to stat %s: %s", dirPath, err)
				return
			}
		} else if !stat.IsDir() {
			l.Error().Printf("%s is not a directory", dirPath)
			return
		}

		file, err := os.Create(path.Join(MySQLHomeDir, MySQLAutodumpOutputDir, now+".sql.gz"))
		if err != nil {
			l.Error().Printf("failed to create mysqlautodump output file: %s", err)
			return
		}
		defer func() {
			err := file.Close()
			if err != nil {
				l.Error().Printf("failed to close mysqlautodump output file: %s", err)
				return
			}
		}()

		gzipped, err := gzip.NewWriterLevel(file, gzip.BestCompression)
		if err != nil {
			l.Error().Printf("failed to compress mysqlautodump output file: %s", err)
			return
		}

		n, err := gzipped.Write(out)
		if err != nil {
			l.Error().Printf("failed to gzip output: %s", err)
			return
		} else if n != len(out) {
			l.Error().Printf("failed to gzip output: expected %d bytes, got %d", len(out), n)
			return
		}

		if err := gzipped.Close(); err != nil {
			l.Error().Printf("failed to gzip output: %s", err)
			return
		}

		l.Info().Printf("successfully dump to %s", fullPath)
	}
}

func main() {
	err := gogger.InitFromEnv()
	if err != nil {
		log.Fatalf("failed to init logger from env: %s", err)
		return
	}

	if slices.Contains(os.Args, "--dumpnow") {
		l.Info().Printf("running dumpnow")
		NewDumper()()
		return
	}

	runner := cron.New()

	l.Info().Printf("add cron job for %s", MySQLAutodumpCron)
	_, err = runner.AddFunc(MySQLAutodumpCron, NewDumper())
	if err != nil {
		l.Error().Fatalf("failed to add cron job: %s", err)
		return
	}

	go func() {
		l.Info().Printf("starting cron")
		runner.Start()
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	l.Info().Printf("press ctrl-c to stop")
	<-sigs

	l.Info().Printf("shutting down")
	runner.Stop()
	l.Info().Printf("shutdown complete")
}
