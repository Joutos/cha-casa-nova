package database

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/chadecasanova/backend/migrations"
)

func TestMigrateIsIdempotentAndPersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "app.db")
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		db, err := Open(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := Migrate(ctx, db, migrations.FS); err != nil {
			t.Fatalf("rodada %d: %v", i, err)
		}
		if i == 0 {
			if err := db.Exec(`INSERT INTO guests (id,name,email,created_at,updated_at) VALUES ('1','Ana','a@b.com',datetime('now'),datetime('now'))`).Error; err != nil {
				t.Fatal(err)
			}
		} else {
			var n int64
			db.Raw("SELECT COUNT(*) FROM guests").Scan(&n)
			if n != 1 {
				t.Fatal("dados não persistiram entre aberturas")
			}
			var v int64
			db.Raw("SELECT COUNT(*) FROM schema_migrations").Scan(&v)
			if v != 1 {
				t.Fatalf("migration registrada %d vezes", v)
			}
			// Integridade referencial ativa.
			if err := db.Exec(`INSERT INTO messages (id,guest_id,body,created_at) VALUES ('m','inexistente','oi',datetime('now'))`).Error; err == nil {
				t.Fatal("chave estrangeira deveria impedir mensagem órfã")
			}
			// E-mail único sem diferenciar maiúsculas.
			if err := db.Exec(`INSERT INTO guests (id,name,email,created_at,updated_at) VALUES ('2','Ana','A@B.COM',datetime('now'),datetime('now'))`).Error; err == nil {
				t.Fatal("e-mail duplicado deveria ser rejeitado")
			}
		}
		s, _ := db.DB()
		s.Close()
	}
}
