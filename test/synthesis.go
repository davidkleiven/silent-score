package test

import (
	"fmt"
	"time"

	"github.com/davidkleiven/silent-score/internal/db"
	"pgregory.net/rapid"
)

func GenerateProjectContentRecord(t *rapid.T) db.ProjectContentRecord {
	hour := rapid.Int32Range(0, 20).Draw(t, "hour")
	start, err := time.Parse(time.TimeOnly, fmt.Sprintf("%02d:%02d:%02d", hour, hour, hour))
	if err != nil {
		panic(err)
	}

	stringSampler := rapid.StringMatching(`[a-zA-Z0-9 ]*`)
	return db.ProjectContentRecord{
		Project:   nil,
		ProjectID: rapid.Uint().Draw(t, "id"),
		Scene:     rapid.Uint().Draw(t, "scene"),
		SceneDesc: stringSampler.Draw(t, "text"),
		Start:     start,
		Keywords:  stringSampler.Draw(t, "keywords"),
		Tempo:     rapid.UintMax(200).Draw(t, "tempo"),
		Theme:     rapid.UintMax(20).Draw(t, "theme"),
	}
}

// Generate a complete project records with a reference to a materialized project
func GenerateCompleteProjectRecords(t *rapid.T) []db.ProjectContentRecord {
	numProjects := rapid.IntRange(0, 3).Draw(t, "numProjects")
	stringSampler := rapid.StringMatching(`[a-zA-Z0-9]+`)
	projectNames := rapid.SliceOfNDistinct(stringSampler, numProjects, numProjects, func(x string) string { return x }).Draw(t, "names")
	numRecordsPerProject := rapid.SliceOfN(rapid.IntRange(0, 5), numProjects, numProjects).Draw(t, "numPerProject")

	totNumRecords := 0
	for _, n := range numRecordsPerProject {
		totNumRecords += n
	}

	recordGen := rapid.Custom(GenerateProjectContentRecord)
	records := rapid.SliceOfN(recordGen, totNumRecords, totNumRecords).Draw(t, "records")
	start := 0
	for i := range numProjects {
		project := db.Project{
			Name: projectNames[i],
		}
		for j := range numRecordsPerProject[i] {
			records[start+j].Project = &project
			records[start+j].Scene = uint(j)
		}
		start += numRecordsPerProject[i]
	}
	return records
}
