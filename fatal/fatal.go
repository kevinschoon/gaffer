package fatal

import (
	"github.com/mesanine/gaffer/log"
	"os"
)

const msg = `
#########################################################################################
#########################################################################################
#########################################################################################
############# FATAL SYSTEM ERROR. EVERYTHING IS BROKEN. TAKE THE DAY OFF. ###############
#########################################################################################
#########################################################################################
#########################################################################################
https://blog.acolyer.org/2017/06/15/gray-failure-the-achilles-heel-of-cloud-scale-systems
`

// FailHard is a global variable that determines
// if Gaffer should fail gradually or all or all
// at once.
var FailHard bool

// Fatal will cause a kernel panic
// if FailHard is true.
func Fatal() {
	if FailHard {
		log.Log.Error(msg)
		if os.Getuid() == 0 {
			fd, _ := os.OpenFile("/proc/sysrq-trigger", os.O_WRONLY, 0)
			if fd != nil {
				// No going back now!
				fd.Write([]byte("c"))
			}
		}
		os.Exit(1)
	}
}
