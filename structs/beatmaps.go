package structs

type DBeatmap struct {
	ID      int
	Title   string
	Status  string
	Version string
	MD5     string
}

type BanchoBeatmap struct {
	Updated    string `json:"last_update"`
	RankStatus string `json:"approved"`
}
