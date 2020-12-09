// +build !windows

package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"goPanel/src/common"
	"goPanel/src/constants"
	"io/ioutil"
	"strconv"
	"syscall"
)

var RestartCmd = cli.Command{
	Name:        "restart",
	Usage:       "restart service gpc",
	Description: "Go panel client service startup",
	Action:      restartRun,
	Flags:       []cli.Flag{},
}

func restartRun(c *cli.Context) {
	pidStr, err := ioutil.ReadFile(common.GetCurrentDir() + constants.GPC_PID_PATH + constants.PID_FILENAME)
	if err != nil {
		log.Error(err)
	}

	pid, err := strconv.Atoi(string(pidStr))
	if err != nil {
		log.Panic(err)
	}

	_ = syscall.Kill(pid, syscall.SIGUSR2)
}
