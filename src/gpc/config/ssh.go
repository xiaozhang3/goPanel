package config

import (
	log "github.com/sirupsen/logrus"
	"goPanel/src/common"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

var (
	sshConfigPath        string
	sshConfigFileName    string
	exampleSshConfigPath string
)

type SshConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Port     int    `yaml:"port"`
}

func (c *SshConfig) initialization() {
	sshConfigPath = Conf.App.UserDir + "/.config/"
	sshConfigFileName = "gpc.yaml"
	exampleSshConfigPath = common.GetCurrentDir() + "/script/client.gpc.yaml.example"

	if !common.DirOrFileByIsExists(sshConfigPath) {
		if !common.CreatePath(sshConfigPath) {
			log.Panic("目录创建失败！")
		}
	}

	sshConfigPathFileName := sshConfigPath + sshConfigFileName
	if !common.DirOrFileByIsExists(sshConfigPathFileName) {
		fileData, err := ioutil.ReadFile(exampleSshConfigPath)
		if err != nil {
			log.Panic("默认配置文件不存在！#", err)
		}

		fp, err := os.Create(sshConfigPathFileName)
		if err != nil {
			log.Panic("文件创建失败！", err)
		}

		if err = ioutil.WriteFile(sshConfigPathFileName, fileData, 0755); err != nil {
			log.Panic("ssh配置文件写入失败！", err)
		}

		defer fp.Close()
	}

	yamlFile, err := ioutil.ReadFile(sshConfigPathFileName)
	if err != nil {
		log.Panic("yamlFile.Get err #%v ", err)
	}

	if err = yaml.Unmarshal(yamlFile, Conf); err != nil {
		log.Panic("Unmarshal: %v", err)
	}
}
