package archive

import (
	"encoding/json"
	"os"
	"sync"
)

type Archiver struct {
	Filename string
	lock     sync.Mutex
}

func (a *Archiver) Store(thing any) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	f, err := os.OpenFile(a.Filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	data, err := json.Marshal(thing)
	if err != nil {
		return err
	}
	data = append(data, []byte(",\n")...)
	_, err = f.Write(data)
	return err

}
