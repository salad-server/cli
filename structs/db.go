package structs

type DBeatmap struct {
	ID      int
	Title   string
	Status  string
	Version string
	MD5     string
}

type Score struct {
	Mode     int
	UserID   int
	MD5      string
	Username string
}
