package common

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"myGin/libs/lib_mongo"
	"os"
	"path"
	"runtime/pprof"
)

// GEnv 全局变量env
var GEnv Env

var (
	cpuProfilingFile,
	memProfilingFile,
	blockProfilingFile,
	goroutineProfilingFile,
	threadCreateProfilingFile *os.File
)

type Env struct {
	Cfg      *Config
	MongoCli lib_mongo.DBAdaptor
}

type Config struct {
	ProjectName string   `yaml:"ProjectName"`
	Log         LogCfg   `yaml:"Log"`
	Mongodb     MongoCfg `yaml:"Mongodb"`
}

type LogCfg struct {
	LogPath   string `yaml:"LogPath"`
	LogLevel  string `yaml:"LogLevel"`
	IsStdOut  bool   `yaml:"IsStdOut"`
	IsPProf   bool   `yaml:"IsPProf"`
	PathPProf string `yaml:"PathPProf"`
}

type MongoCfg struct {
	Host      string `yaml:"Host"`
	User      string `yaml:"User"`
	Passwd    string `yaml:"Passwd"`
	DbName    string `yaml:"DbName"`
	PoolLimit uint64 `yaml:"PoolLimit"`
}

func newEnv() *Env {
	return &GEnv
}

func GetEnv() *Env {
	return &GEnv
}

func EnvBoot(p string) {
	content, err := ioutil.ReadFile(p)
	if err != nil {
		panic(err)
	}
	var c Config
	err = yaml.Unmarshal(content, &c)
	if err != nil {
		panic(err)
	}
	env := newEnv()
	env.Cfg = &c
}

func InitDebugPProf(setting *Config) error {
	if setting.Log.IsPProf {
		_, err := os.Stat(setting.Log.PathPProf)
		if err != nil && os.IsNotExist(err) {
			_ = os.MkdirAll(setting.Log.PathPProf, os.ModePerm)
		}

		pathPrefix := path.Join(setting.Log.PathPProf, fmt.Sprintf("%d", os.Getpid()))
		logrus.Infof("start pprof, and will save to %s", pathPrefix)
		cpuProfilingFile, _ = os.Create(pathPrefix + "-cpu.prof")
		memProfilingFile, _ = os.Create(pathPrefix + "-mem.prof")
		blockProfilingFile, _ = os.Create(pathPrefix + "-block.prof")
		goroutineProfilingFile, _ = os.Create(pathPrefix + "-goroutine.prof")
		threadCreateProfilingFile, _ = os.Create(pathPrefix + "-threadcreat.prof")
		_ = pprof.StartCPUProfile(cpuProfilingFile)
	}
	return nil
}

// SaveProfile try to save pprof into local file
func (e *Env) SaveProfile() {
	if e.Cfg.Log.IsPProf {
		goroutine := pprof.Lookup("goroutine")
		_ = goroutine.WriteTo(goroutineProfilingFile, 1)
		heap := pprof.Lookup("heap")
		_ = heap.WriteTo(memProfilingFile, 1)
		block := pprof.Lookup("block")
		_ = block.WriteTo(blockProfilingFile, 1)
		threadcreate := pprof.Lookup("threadcreate")
		_ = threadcreate.WriteTo(threadCreateProfilingFile, 1)
		pprof.StopCPUProfile()
	}
}
