package malbeep

import (
	"log"
	"sync"

	"github.com/gen2brain/malgo"
)

var (
	context     *malgo.AllocatedContext
	contextUses uint
	contextLock sync.Mutex
)

func initContext() (*malgo.AllocatedContext, error) {
	contextLock.Lock()
	defer contextLock.Unlock()

	if contextUses == 0 {
		var err error
		context, err = malgo.InitContext(nil, malgo.ContextConfig{}, func (msg string) {
            log.Printf("malgo: %s", msg)
        })

		if err != nil {
			return nil, err
		}
	}
	contextUses++

	return context, nil
}

func freeContext() {
	contextLock.Lock()
	defer contextLock.Unlock()

	contextUses--
	if contextUses == 0 {
		context.Uninit()
		context.Free()
	}
}
