package storage

type Config struct {
	EnduroAddress string
	Internal      LocationConfig
	Database      Database
}

type Database struct {
	DSN     string
	Migrate bool
}

type LocationConfig struct {
	Name      string
	Region    string
	Endpoint  string
	PathStyle bool
	Profile   string
	Key       string
	Secret    string
	Token     string
	Bucket    string
}
