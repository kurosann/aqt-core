package notify

import (
	"errors"
	"sync"
)

var (
	notifierDict = map[Way]Notifier{}
	peopleDict   = map[Type]map[Way][]string{}
	nl           sync.RWMutex
)

func AddNotify(notifier Notifier) {
	nl.Lock()
	defer nl.Unlock()

	notifierDict[notifier.Way()] = notifier
}

func AddPerson(typ Type, way Way, user string) {
	if typ == "" {
		typ = Normal
	}
	nl.Lock()
	defer nl.Unlock()

	if _, ok := peopleDict[typ]; !ok {
		peopleDict[typ] = map[Way][]string{}
	}
	peopleDict[typ][way] = append(peopleDict[typ][way], user)
}

type Way string
type Type string

const (
	Mail   Way = "mail"
	Feishu Way = "feishu"
)
const (
	Alert  Type = "alert"
	Debug  Type = "debug"
	Normal Type = "normal"
)

type Notifier interface {
	Way() Way
	Send(title string, data any, to string) error
}

func SendAllByType(typ Type, title string, data any) error {
	nl.RLock()
	defer nl.RUnlock()
	var errs []error
	for w, persons := range peopleDict[typ] {
		if notifier, ok := notifierDict[w]; ok {
			for _, p := range persons {
				if err := notifier.Send(title, data, p); err != nil {
					errs = append(errs, err)
				}
			}
		} else {
			errs = append(errs, errors.New("not found notifier"))
		}
	}
	return errors.Join(errs...)
}

// SendAllByWay 给所有用户发送某一类客户端通知
func SendAllByWay(way Way, title string, data any) error {
	nl.RLock()
	defer nl.RUnlock()
	var errs []error
	if notifier, ok := notifierDict[way]; ok {
		for _, persons := range peopleDict {
			for _, p := range persons[Mail] {
				if err := notifier.Send(title, data, p); err != nil {
					errs = append(errs, err)
				}
			}
		}
	} else {
		return errors.New("not found notifier")
	}
	return errors.Join(errs...)
}
