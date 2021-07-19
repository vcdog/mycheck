package loggs

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

type BaseInfo struct {
	Logs LogsEntity `yaml:"logs"`
	DBinfo DBinfoEntity `yaml:"dbinfo"`
}
type DBinfoEntity struct {
	DirverName string `yaml:"dirverName"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	Database string `yaml:"database"`
	Charset string `yaml:"charset"`
	DBconnIdleTime time.Duration `yaml:"dbConnIdleTime"`
	MaxIdleConns int `yaml:"maxIdleConns"`
}
type LogsEntity struct {
	Loglevel   string        `yaml:"loglevel"`
	OutputFile OutputFileEntity `yaml:"outputFile"`
}
type OutputFileEntity struct {
	Logfile string `yaml: "logfile"`
	LogLevel string `yaml:loglevel"`
	LogMaxAge time.Duration  `yaml:"logMaxAge"`
	IsConsole bool `yaml:"isConsole"`
	RotationTime time.Duration `yaml:"rotationTime"`
	Skip int `yaml:"skip"`
}
func (conf *BaseInfo) GetConf() BaseInfo {
	yamlFile,err := ioutil.ReadFile("D:\\goProject\\DepthInspection\\conf\\logs.yaml")
	if err != nil{
		fmt.Println(err.Error())
	}
	//将读取到的yaml文件解析为响应的struct
	err = yaml.Unmarshal(yamlFile,&conf)
	if err != nil{
		fmt.Println(err.Error())
	}
	return *conf
}




