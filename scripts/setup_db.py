import logging
import os
import sqlite3
import sys

sys.path.append(os.path.dirname(__file__) + "/../dashboard/")

from utils import load_configs_from_file_without_cache

logging.basicConfig(
    encoding="utf-8",
    level=logging.INFO,
    format="%(asctime)s :: %(levelname)-8s :: %(name)s :: %(message)s",
)
logger = logging.getLogger()


class SetupDB:
    def __init__(self) -> None:
        configs = load_configs_from_file_without_cache()
        self.dbs_folder = configs["database"]["databases_folder_abs_path"]

    def create_dbs(self):
        logger.info("Creating databases")
        self.create_trackers_db()
        logger.info("Databases created")

    def create_trackers_db(self):
        db_path = os.path.join(self.dbs_folder, "trackers.db")
        logger.info(f"Creating trackers database at path {db_path}")
        conn = sqlite3.connect(db_path)
        cur = conn.cursor()

        sql = """
CREATE TABLE IF NOT EXISTS games_tracker (
    url VARCHAR(200),
    name VARCHAR(50) PRIMARY KEY,
    cover_img BLOB,
    release_date DATE,
    tags TEXT,
    developers TEXT,
    publishers TEXT,
    priority SMALLINT,
    status SMALLINT,
    stars SMALLINT,
    purchased_or_gamepass BOOLEAN,
    started_date DATE,
    finished_dropped_date DATE,
    commentary TEXT
)
        """
        logger.info("Creating games_tracker table")
        cur.execute(sql)

        sql = """
CREATE TABLE IF NOT EXISTS medias_tracker (
    url VARCHAR(200),
    name VARCHAR(50) PRIMARY KEY,
    media_type VARCHAR(20),
    cover_img BLOB,
    release_date DATE,
    genres TEXT,
    staff TEXT,
    priority SMALLINT,
    status SMALLINT,
    stars SMALLINT,
    started_date DATE,
    finished_dropped_date DATE,
    commentary TEXT
)
        """
        logger.info("Creating medias_tracker table")
        cur.execute(sql)

        conn.commit()


def main():
    setup = SetupDB()
    setup.create_dbs()


if __name__ == "__main__":
    main()
