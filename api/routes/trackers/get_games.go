package trackers

import (
	"database/sql"
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/diogovalentte/dashboard/api/util"
	"github.com/gin-gonic/gin"
)

func GetAllGames(c *gin.Context) {
	sqlQuery := fmt.Sprintf(`
SELECT
  url, name, cover_img, release_date, tags, developers, publishers, priority,
  status, stars, purchased_or_gamepass, started_date, finished_dropped_date, commentary
FROM
  games_tracker;`,
	)

	games, err := getGamesFromQuery(sqlQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"games": games})
}

func GetPlayingGames(c *gin.Context) {
	sqlQuery := fmt.Sprintf(`
SELECT
  url, name, cover_img, release_date, tags, developers, publishers, priority,
  status, stars, purchased_or_gamepass, started_date, finished_dropped_date, ""
FROM
  games_tracker
WHERE
  status = 3
ORDER BY
  started_date DESC;`,
	)

	games, err := getGamesFromQuery(sqlQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}

	// Sort games by the Started Date putting the nulls at the end
	var nullDate time.Time
	sort.Slice(games, func(i, j int) bool {
		if games[i].StartedDate == nullDate {
			return false
		}
		if games[j].StartedDate == nullDate {
			return true
		}

		return (*games[i]).StartedDate.After((*games[j]).StartedDate)
	})

	c.JSON(http.StatusOK, gin.H{"games": games})
}

func GetToBeReleasedGames(c *gin.Context) {
	sqlQuery := fmt.Sprintf(`
SELECT
  url, name, cover_img, release_date, tags, developers, publishers, priority,
  status, stars, purchased_or_gamepass, started_date, finished_dropped_date, ""
FROM
  games_tracker
WHERE
  status = 1
ORDER BY
  release_date;`,
	)

	games, err := getGamesFromQuery(sqlQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}

	// Sort games by the release date putting the nulls at the end
	var nullDate time.Time
	sort.Slice(games, func(i, j int) bool {
		if games[i].ReleaseDate == nullDate {
			return false
		}
		if games[j].ReleaseDate == nullDate {
			return true
		}

		return (*games[j]).ReleaseDate.After((*games[i]).ReleaseDate)
	})

	c.JSON(http.StatusOK, gin.H{"games": games})
}

func GetNotStartedGames(c *gin.Context) {
	sqlQuery := fmt.Sprintf(`
SELECT
  url, name, cover_img, release_date, tags, developers, publishers, priority,
  status, stars, purchased_or_gamepass, started_date, finished_dropped_date, ""
FROM
  games_tracker
WHERE
  status = 2
ORDER BY
  CASE
    WHEN priority = 1 THEN 1
    WHEN priority = 2 THEN 2
    WHEN priority = 3 THEN 3
  END;`,
	)

	games, err := getGamesFromQuery(sqlQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"games": games})
}

func GetFinishedGames(c *gin.Context) {
	sqlQuery := fmt.Sprintf(`
SELECT
  url, name, cover_img, release_date, tags, developers, publishers, priority,
  status, stars, purchased_or_gamepass, started_date, finished_dropped_date, ""
FROM
  games_tracker
WHERE
  status = 4
ORDER BY
  finished_dropped_date DESC;`,
	)

	games, err := getGamesFromQuery(sqlQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}

	// Sort games by the Finished/Dropped date putting the nulls at the end
	var nullDate time.Time
	sort.Slice(games, func(i, j int) bool {
		if games[i].FinishedDroppedDate == nullDate {
			return false
		}
		if games[j].FinishedDroppedDate == nullDate {
			return true
		}

		return (*games[i]).FinishedDroppedDate.After((*games[j]).FinishedDroppedDate)
	})

	c.JSON(http.StatusOK, gin.H{"games": games})
}

func GetDroppedGames(c *gin.Context) {
	sqlQuery := fmt.Sprintf(`
SELECT
  url, name, cover_img, release_date, tags, developers, publishers, priority,
  status, stars, purchased_or_gamepass, started_date, finished_dropped_date, ""
FROM
  games_tracker
WHERE
  status = 5
ORDER BY
  finished_dropped_date DESC;`,
	)

	games, err := getGamesFromQuery(sqlQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}

	// Sort games by the Finished/Dropped date putting the nulls at the end
	var nullDate time.Time
	sort.Slice(games, func(i, j int) bool {
		if games[i].FinishedDroppedDate == nullDate {
			return false
		}
		if games[j].FinishedDroppedDate == nullDate {
			return true
		}

		return (*games[i]).FinishedDroppedDate.After((*games[j]).FinishedDroppedDate)
	})

	c.JSON(http.StatusOK, gin.H{"games": games})
}

func getGamesFromQuery(sqlQuery string) ([]*GameProperties, error) {
	configs, err := util.GetConfigsWithoutDefaults("../../../configs/")
	if err != nil {
		return nil, err
	}
	dbPath := filepath.Join(configs.Database.FolderPath, "trackers.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(sqlQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gamesProperties []*GameProperties
	for rows.Next() {
		gameProperties := GameProperties{}
		var tagsStr string
		var developersStr string
		var publishersStr string
		err = rows.Scan(
			&gameProperties.URL,
			&gameProperties.Name,
			&gameProperties.CoverImg,
			&gameProperties.ReleaseDate,
			&tagsStr,
			&developersStr,
			&publishersStr,
			&gameProperties.Priority,
			&gameProperties.Status,
			&gameProperties.Stars,
			&gameProperties.PurchasedOrGamePass,
			&gameProperties.StartedDate,
			&gameProperties.FinishedDroppedDate,
			&gameProperties.Commentary)
		if err != nil {
			return nil, err
		}
		gameProperties.Tags = strings.Split(tagsStr, ",")
		gameProperties.Developers = strings.Split(developersStr, ",")
		gameProperties.Publishers = strings.Split(publishersStr, ",")

		gamesProperties = append(gamesProperties, &gameProperties)
	}

	return gamesProperties, nil
}
