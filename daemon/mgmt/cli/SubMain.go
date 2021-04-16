/**
 * @Author: wzx
 * @Description:
 * @Version: 1.0.0
 * @Date: 2021/3/30 下午7:13
 *@Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
 */

package mgmt

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func SubMain() {
	oApp := cli.NewApp()
	oApp.Name = "mirc"
	oApp.Usage = " MIR Management Cli Tools "
	oApp.Commands = []*cli.Command{
		&faceCommands,
	}

	if err := oApp.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
