package opts

import "github.com/portey/image-resizer/storage/minio"

type Config struct {
	PrettyLogOutput bool
	LogLevel        string

	GraphQLPort     int
	HealthCHeckPort int

	MongoURI      string
	MongoDatabase string

	StorageCfg minio.Config
}
