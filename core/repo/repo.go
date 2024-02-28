package repo

import "common/database"

type Manager struct {
	Mongo *database.MongoManager
	Redis *database.RedisManager
}

func New() *Manager {
	return &Manager{
		database.NewMongo(),
		database.NewRedis(),
	}
}
