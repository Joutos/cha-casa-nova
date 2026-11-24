// Package database abre a conexão SQLite e aplica as migrations versionadas.
package database

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open abre o banco com chaves estrangeiras ligadas, WAL e timeout de lock.
// path pode ser um caminho de arquivo ou ":memory:" (testes).
func Open(path string) (*gorm.DB, error) {
	var dsn string
	if path == ":memory:" {
		dsn = "file::memory:?_foreign_keys=on&_busy_timeout=5000"
	} else {
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			return nil, fmt.Errorf("criando diretório do banco: %w", err)
		}
		dsn = fmt.Sprintf("file:%s?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&_txlock=immediate", path)
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger:                 logger.Discard,
		NowFunc:                func() time.Time { return time.Now().UTC() },
		SkipDefaultTransaction: true,
		TranslateError:         true,
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// SQLite aceita um escritor por vez; uma única conexão elimina SQLITE_BUSY
	// e é mais do que suficiente para o volume deste site.
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetConnMaxLifetime(0)
	return db, nil
}

// Migrate aplica, em ordem, todas as migrations *.up.sql ainda não registradas.
func Migrate(ctx context.Context, db *gorm.DB, migrations fs.FS) error {
	if err := db.WithContext(ctx).Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL
	)`).Error; err != nil {
		return err
	}

	files, err := fs.Glob(migrations, "*.up.sql")
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, name := range files {
		version := strings.TrimSuffix(name, ".up.sql")
		var count int64
		if err := db.WithContext(ctx).Table("schema_migrations").Where("version = ?", version).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		body, err := fs.ReadFile(migrations, name)
		if err != nil {
			return err
		}
		err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(string(body)).Error; err != nil {
				return fmt.Errorf("migration %s: %w", name, err)
			}
			return tx.Exec("INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)", version, time.Now().UTC()).Error
		})
		if err != nil {
			return err
		}
		slog.Info("migration aplicada", "version", version)
	}
	return nil
}
