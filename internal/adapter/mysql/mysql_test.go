package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	mysqltc "github.com/testcontainers/testcontainers-go/modules/mysql"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := mysqltc.Run(ctx, "mysql:8.0",
		mysqltc.WithDatabase("task_vault_test"),
		mysqltc.WithUsername("test"),
		mysqltc.WithPassword("test"),
	)
	if err != nil {
		fmt.Printf("не удалось запустить контейнер: %v\n", err)
		os.Exit(1)
	}
	defer container.Terminate(ctx)

	dsn, err := container.ConnectionString(ctx, "parseTime=true", "multiStatements=true")
	if err != nil {
		fmt.Printf("не удалось получить DSN: %v\n", err)
		os.Exit(1)
	}

	testDB, err = sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("не удалось подключиться: %v\n", err)
		os.Exit(1)
	}
	defer testDB.Close()

	if err := runMigrations(testDB); err != nil {
		fmt.Printf("ошибка миграций: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func runMigrations(db *sql.DB) error {
	driver, err := migratemysql.WithInstance(db, &migratemysql.Config{})
	if err != nil {
		return err
	}

	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "..", "..", "migrations")
	absPath, _ := filepath.Abs(migrationsPath)

	m, err := migrate.NewWithDatabaseInstance("file://"+absPath, "mysql", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func cleanTables(t *testing.T) {
	t.Helper()
	tables := []string{"task_comments", "task_history", "tasks", "team_members", "teams", "users"}
	for _, table := range tables {
		_, err := testDB.Exec("DELETE FROM " + table)
		if err != nil {
			t.Fatalf("не удалось очистить таблицу %s: %v", table, err)
		}
	}
}
