package bootstrap

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"myGin/common"
	"myGin/libs/lib_mongo"
	"os"
	"path/filepath"
)

func Bootstrap(p string) error {
	common.EnvBoot(p)
	env := common.GetEnv()
	if err := InitLog(env.Cfg); err != nil {
		return err
	}
	if err := common.InitDebugPProf(env.Cfg); err != nil {
		return err
	}
	mongoClient, err := InitMongoClient(env.Cfg)
	if err != nil {
		return err
	}
	env.MongoCli = mongoClient
	return nil
}

func InitLog(setting *common.Config) error {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if setting.Log.LogLevel == "" {
		return nil
	}
	lvl, err := logrus.ParseLevel(setting.Log.LogLevel)
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)

	if setting.Log.IsStdOut {
		logrus.SetOutput(os.Stdout)
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
	if setting.Log.LogPath != "" {
		logFile := filepath.Join(setting.Log.LogPath, setting.ProjectName+"_stdout.log")
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			logrus.SetOutput(file)
		}
	}
	logrus.SetReportCaller(true)
	return nil
}

func InitMongoClient(setting *common.Config) (lib_mongo.DBAdaptor, error) {
	mongoCli := lib_mongo.NewMongoSession()
	MongoURL := fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=%s",
		setting.Mongodb.User, setting.Mongodb.Passwd, setting.Mongodb.Host, setting.Mongodb.DbName, setting.Mongodb.DbName)
	err := mongoCli.Connect(MongoURL, setting.Mongodb.DbName)
	if err != nil {
		return nil, err
	}
	mongoCli.SetPoolLimit(setting.Mongodb.PoolLimit)
	return mongoCli, nil
}
