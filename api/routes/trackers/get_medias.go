package trackers

import (
	"database/sql"
	"fmt"
	"github.com/diogovalentte/dashboard/api/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func GetAllMedias(c *gin.Context) {
	sqlQuery := fmt.Sprintf(`
SELECT
  url, name, media_type, cover_img, release_date, genres, staff,
  priority, status, stars, started_date, finished_dropped_date
FROM
  medias_tracker;`,
	)

	medias, err := getMediasFromQuery(sqlQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"medias": medias})
}

func GetWatchingReadingMedias(c *gin.Context) {
	sqlQuery := fmt.Sprintf(`
SELECT
  url, name, media_type, cover_img, release_date, genres, staff,
  priority, status, stars, started_date, finished_dropped_date
FROM
  medias_tracker
WHERE
  status = 'Watching/Reading'
ORDER BY
  started_date DESC;`,
	)

	medias, err := getMediasFromQuery(sqlQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}

	// Sort medias by the Started date putting the nulls at the end
	var nullDate time.Time
	sort.Slice(medias, func(i, j int) bool {
		if medias[i].StartedDate == nullDate {
			return false
		}
		if medias[j].StartedDate == nullDate {
			return true
		}

		return (*medias[i]).StartedDate.After((*medias[j]).StartedDate)
	})

	c.JSON(http.StatusOK, gin.H{"medias": medias})
}

func GetToBeReleasedMedias(c *gin.Context) {
	sqlQuery := fmt.Sprintf(`
SELECT
  url, name, media_type, cover_img, release_date, genres, staff,
  priority, status, stars, started_date, finished_dropped_date
FROM
  medias_tracker
WHERE
  status = 'To be released'
ORDER BY
  release_date;`,
	)

	medias, err := getMediasFromQuery(sqlQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}

	// Sort medias by the release date putting the nulls at the end
	var nullDate time.Time
	sort.Slice(medias, func(i, j int) bool {
		if medias[i].ReleaseDate == nullDate {
			return false
		}
		if medias[j].ReleaseDate == nullDate {
			return true
		}

		return (*medias[j]).ReleaseDate.After((*medias[i]).ReleaseDate)
	})

	c.JSON(http.StatusOK, gin.H{"medias": medias})
}

func GetNotStartedMedias(c *gin.Context) {
	sqlQuery := fmt.Sprintf(`
SELECT
  url, name, media_type, cover_img, release_date, genres, staff,
  priority, status, stars, started_date, finished_dropped_date
FROM
  medias_tracker
WHERE
  status = 'Not started'
ORDER BY
  CASE
    WHEN priority = 'High' THEN 1
    WHEN priority = 'Medium' THEN 2
    WHEN priority = 'Low' THEN 3
  END;`,
	)

	medias, err := getMediasFromQuery(sqlQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"medias": medias})
}

func GetFinishedMedias(c *gin.Context) {
	sqlQuery := fmt.Sprintf(`
SELECT
  url, name, media_type, cover_img, release_date, genres, staff,
  priority, status, stars, started_date, finished_dropped_date
FROM
  medias_tracker
WHERE
  status = 'Finished'
ORDER BY
  finished_dropped_date DESC;`,
	)

	medias, err := getMediasFromQuery(sqlQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}

	// Sort medias by the Finished/Dropped date putting the nulls at the end
	var nullDate time.Time
	sort.Slice(medias, func(i, j int) bool {
		if medias[i].FinishedDroppedDate == nullDate {
			return false
		}
		if medias[j].FinishedDroppedDate == nullDate {
			return true
		}

		return (*medias[i]).FinishedDroppedDate.After((*medias[j]).FinishedDroppedDate)
	})

	c.JSON(http.StatusOK, gin.H{"medias": medias})
}

func GetDroppedMedias(c *gin.Context) {
	sqlQuery := fmt.Sprintf(`
SELECT
  url, name, media_type, cover_img, release_date, genres, staff,
  priority, status, stars, started_date, finished_dropped_date
FROM
  medias_tracker
WHERE
  status = 'Dropped'
ORDER BY
  finished_dropped_date DESC;`,
	)

	medias, err := getMediasFromQuery(sqlQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}

	// Sort medias by the Finished/Dropped date putting the nulls at the end
	var nullDate time.Time
	sort.Slice(medias, func(i, j int) bool {
		if medias[i].FinishedDroppedDate == nullDate {
			return false
		}
		if medias[j].FinishedDroppedDate == nullDate {
			return true
		}

		return (*medias[i]).FinishedDroppedDate.After((*medias[j]).FinishedDroppedDate)
	})

	c.JSON(http.StatusOK, gin.H{"medias": medias})
}

func getMediasFromQuery(sqlQuery string) ([]*MediaProperties, error) {
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

	var mediasProperties []*MediaProperties
	for rows.Next() {
		mediaProperties := MediaProperties{}
		var genresStr string
		var staffStr string
		err = rows.Scan(
			&mediaProperties.URL,
			&mediaProperties.Name,
			&mediaProperties.MediaType,
			&mediaProperties.CoverImg,
			&mediaProperties.ReleaseDate,
			&genresStr,
			&staffStr,
			&mediaProperties.Priority,
			&mediaProperties.Status,
			&mediaProperties.Stars,
			&mediaProperties.StartedDate,
			&mediaProperties.FinishedDroppedDate)
		if err != nil {
			return nil, err
		}
		mediaProperties.Genres = strings.Split(genresStr, ",")
		mediaProperties.Staff = strings.Split(staffStr, ",")

		mediasProperties = append(mediasProperties, &mediaProperties)
	}

	return mediasProperties, nil
}
