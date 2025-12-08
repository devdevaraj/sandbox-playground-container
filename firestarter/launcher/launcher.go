package launcher

import (
	"context"
	"log"
	"strconv"

	"github.com/devdevaraj/firestarter/init_app"
)

func Launcher(cfg init_app.Config, ctx context.Context) {
	ConfigureNetWork(cfg)
	for i := range len(cfg.Templates) {
		go func() {
			StartMicroVM(
				ctx,
				"vm"+strconv.Itoa(i+1),
				cfg.Templates[i].Disks,
				cfg.Templates[i].Network,
				cfg.Templates[i].KernelArgs,
				cfg.Templates[i].Kernel,
				cfg.Templates[i].RootFS,
				cfg.Templates[i].CPU,
				cfg.Templates[i].SMT,
				cfg.Templates[i].RAM,
				cfg.Templates[i].EnableOverlay,
				cfg.Templates[i].IsZFS,
				"/"+cfg.Templates[i].ZFSClonePath+"/"+cfg.Name+"-vm"+strconv.Itoa(i+1)+"/rootfs.ext4",
			)
			log.Printf("After starting VM")
		}()
	}
}
