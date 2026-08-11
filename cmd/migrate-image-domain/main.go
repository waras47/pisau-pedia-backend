// migrate-image-domain rewrites every stored URL that starts with an old
// image-storage domain (e.g. the default Cloudflare R2 pub-*.r2.dev domain)
// to a new domain (e.g. a custom domain attached to the same bucket).
//
// It scans every text/JSON column in the database rather than a hardcoded
// table list, so it can't miss a column. Run without -apply first to see
// what would change; nothing is written until -apply is passed.
//
// Usage:
//
//	go run ./cmd/migrate-image-domain -old https://pub-xxxx.r2.dev -new https://cdn.pisaupedia.my.id
//	go run ./cmd/migrate-image-domain -old https://pub-xxxx.r2.dev -new https://cdn.pisaupedia.my.id -apply
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/config"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/database"
)

type column struct {
	Table  string `db:"TABLE_NAME"`
	Column string `db:"COLUMN_NAME"`
}

func main() {
	oldDomain := flag.String("old", "", "old domain prefix to replace, e.g. https://pub-xxxx.r2.dev")
	newDomain := flag.String("new", "", "new domain to replace it with, e.g. https://cdn.pisaupedia.my.id")
	apply := flag.Bool("apply", false, "actually write the changes (default is dry-run)")
	flag.Parse()

	if *oldDomain == "" || *newDomain == "" {
		log.Fatal("both -old and -new are required")
	}
	oldTrimmed := strings.TrimRight(*oldDomain, "/")
	newTrimmed := strings.TrimRight(*newDomain, "/")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	db, err := database.Connect(cfg.DB.DSN())
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	var cols []column
	err = db.Select(&cols, `
		SELECT TABLE_NAME, COLUMN_NAME
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND DATA_TYPE IN ('varchar', 'char', 'text', 'tinytext', 'mediumtext', 'longtext', 'json')
		ORDER BY TABLE_NAME, ORDINAL_POSITION
	`)
	if err != nil {
		log.Fatalf("list columns: %v", err)
	}

	needle := "%" + oldTrimmed + "%"
	totalRows := 0
	touchedCols := 0

	for _, c := range cols {
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `%s` WHERE `%s` LIKE ?", c.Table, c.Column)
		var n int
		if err := db.Get(&n, countQuery, needle); err != nil {
			log.Printf("skip %s.%s: %v", c.Table, c.Column, err)
			continue
		}
		if n == 0 {
			continue
		}

		touchedCols++
		totalRows += n
		fmt.Printf("%s.%s: %d row(s) contain the old domain\n", c.Table, c.Column, n)

		if *apply {
			updateQuery := fmt.Sprintf("UPDATE `%s` SET `%s` = REPLACE(`%s`, ?, ?) WHERE `%s` LIKE ?", c.Table, c.Column, c.Column, c.Column)
			res, err := db.Exec(updateQuery, oldTrimmed, newTrimmed, needle)
			if err != nil {
				log.Printf("  FAILED to update %s.%s: %v", c.Table, c.Column, err)
				continue
			}
			affected, _ := res.RowsAffected()
			fmt.Printf("  -> updated %d row(s)\n", affected)
		}
	}

	fmt.Println()
	if touchedCols == 0 {
		fmt.Println("No rows found containing the old domain. Nothing to do.")
		return
	}
	if *apply {
		fmt.Printf("Done. %d row(s) across %d column(s) updated.\n", totalRows, touchedCols)
	} else {
		fmt.Printf("Dry run: %d row(s) across %d column(s) would be updated. Re-run with -apply to write changes.\n", totalRows, touchedCols)
	}
}
