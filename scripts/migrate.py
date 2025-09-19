import os
import sqlite3
from pathlib import Path


def autodiscover_migrations() -> list[Path]:
    migrations = Path().parent.absolute() / "migrations"
    if not migrations.exists():
        raise FileNotFoundError("[Aborted] missing migrations folder")
    sql_files = [*migrations.glob("*.sql")]
    if len(sql_files) < 1:
        raise FileNotFoundError("[Aborted] no migrations found!")
    return sql_files


def migrate_sqlite(filename: str, files: list[Path]):
    with sqlite3.connect(filename) as instance:
        for file in files:
            print(f"[Info] Running transaction for file {file.name}", flush=True)
            cursor = instance.cursor()
            cursor.executescript(file.read_text())
            cursor.close()
        print("[Info] Committing transaction", flush=True)
        instance.commit()


def main():
    print("[Info] Starting migration...", flush=True)
    files = autodiscover_migrations()
    db = os.getenv("DB_TYPE", "sqlite")
    params = os.getenv("DB_PARAMS", "test.db")
    match db:
        case "sqlite":
            migrate_sqlite(params, files)
    print("[Info] Done", flush=True)


if __name__ == "__main__":
    main()
