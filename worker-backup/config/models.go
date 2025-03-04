package config

type Environment struct {
	MongoHost     string `env:"MONGO_HOST"`
	MongoPort     string `env:"MONGO_HOST,default=27017"`
	MongoUser     string `env:"MONGO_USER"`
	MongoPassword string `env:"MONGO_PASSWORD"`
	MongoDB       string `env:"MONGO_DB"`
	BackupDir     string `env:"BACKUP_DIR"`
	ZipPassword   string `env:"ZIP_PASSWORD"`
	BucketName    string `env:"BUCKET_NAME"`
}
