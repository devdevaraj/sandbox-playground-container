package launcher

import (
	"context"
	"strconv"

	"github.com/devdevaraj/firestarter/init_app"
)

func Launcher(cfg init_app.Config, ctx context.Context, args []string) {
	ConfigureNetWork("172.16.0.1/24", "172.16.0.0/24", len(cfg.Templates))
	for i := range len(cfg.Templates) {
		go StartMicroVM(
			ctx,
			"vm"+strconv.Itoa(i+1),
			"tap"+strconv.Itoa(i),
			"172.16.0."+strconv.Itoa(i+2),
			"172.16.0.1",
			"AA:FC:00:00:00:0"+strconv.Itoa(i+1),
			cfg.Templates[i].KernelArgs,
			cfg.Templates[i].Kernel,
			cfg.Templates[i].RootFS,
			cfg.Templates[i].CPU,
			cfg.Templates[i].SMT,
			cfg.Templates[i].RAM,
			cfg.Templates[i].EnableOverlay,
			cfg.Templates[i].IsZFS,
			"/"+cfg.Templates[i].ZFSClonePath+"/"+args[2]+"-vm"+strconv.Itoa(i+1)+"/rootfs.ext4",
		)
	}
}
