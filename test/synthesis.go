package test

import (
	"github.com/davidkleiven/silent-score/internal/db"
	"pgregory.net/rapid"
)

func GenerateProjectContentRecord(t *rapid.T, projectId uint) db.ProjectContentRecord {
	stringSampler := rapid.StringMatching(`[a-zA-Z0-9 ]*`)
	return db.ProjectContentRecord{
		ProjectID:   projectId,
		Scene:       rapid.Uint().Draw(t, "scene"),
		SceneDesc:   stringSampler.Draw(t, "text"),
		DurationSec: rapid.IntRange(0, 100).Draw(t, "duration"),
		Keywords:    stringSampler.Draw(t, "keywords"),
		Tempo:       rapid.UintMax(200).Draw(t, "tempo"),
		Theme:       rapid.UintMax(20).Draw(t, "theme"),
	}
}

// Generate a complete project records with a reference to a materialized project
func GenerateProjects(t *rapid.T) []db.Project {
	numProjects := rapid.IntRange(1, 3).Draw(t, "numProjects")
	stringSampler := rapid.StringMatching(`[a-zA-Z0-9]+`)
	projectNames := rapid.SliceOfNDistinct(stringSampler, numProjects, numProjects, func(x string) string { return x }).Draw(t, "names")
	numRecordsPerProject := rapid.SliceOfN(rapid.IntRange(0, 5), numProjects, numProjects).Draw(t, "numPerProject")

	projects := make([]db.Project, numProjects)
	for i := range numProjects {
		id := uint(i + 1)
		recordGen := rapid.Custom(func(t *rapid.T) db.ProjectContentRecord { return GenerateProjectContentRecord(t, id) })
		projects[i] = db.Project{
			Name:    projectNames[i],
			Id:      id,
			Records: rapid.SliceOfN(recordGen, numRecordsPerProject[i], numRecordsPerProject[i]).Draw(t, "records"),
		}
	}

	return projects
}
