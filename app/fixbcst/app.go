package fixbcst

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nats-io/go-nats"
	"github.com/tkanos/gonfig"
	"golang.org/x/sys/windows/svc"
)

// SubjClass for Publish
var SubjClass = "DST.ETS.MD"
var ConfigFile = "config.json"

// FixBcstConfiguration
type FixBcstConfiguration struct {
	EtsConnString  string
	NatsConnString string
	SubjectPublish string
	LogFileName    string
	LogTimeFlag    bool
	LogDebugFlag   bool
	LogTraceFlag   bool
	LogSysLog      bool
}

// Logging
type Logging struct {
	sync.RWMutex
	logger Logger
	trace  int32
	debug  int32
}

// FixBcst main struct
type FixBcst struct {
	config  FixBcstConfiguration
	logging Logging
}

func NewFixBcst() *FixBcst {
	config := FixBcstConfiguration{}
	workingPath := filepath.Dir(os.Args[0])
	configPath := filepath.Join(workingPath, ConfigFile)
	err := gonfig.GetConf(configPath, &config)
	if err != nil {
		log.Fatalln(err)
	}
	// Override Subject to Publish
	if config.SubjectPublish != "" {
		SubjClass = config.SubjectPublish
	}
	if config.LogFileName != "" {
		// Add to exec path
		config.LogFileName = filepath.Join(workingPath, config.LogFileName)
	}

	s := &FixBcst{
		config: config,
	}
	s.ConfigureLogger()

	return s
}

func (s *FixBcst) StartUp() {
	//s.winlog.Info(1, "FixBcstApp is connectiong to ETS.")
	ct := CtciNew(s.config.EtsConnString, s.logging)
	err := ct.Dial()
	if err != nil {
		//s.winlog.Error(1002, err.Error())
		s.logging.logger.Fatalf("%v\n", err)
	}
	defer ct.Close()

	// NATS
	//DefaultURL              = "nats://localhost:4222"
	var urls = flag.String("s", s.config.NatsConnString, "The nats server URLs (separated by comma)")

	//s.winlog.Info(1, "FixBcstApp is connectiong to NATS.")
	nc, err := nats.Connect(*urls)
	if err != nil {
		//s.winlog.Error(1003, err.Error())
		s.logging.logger.Fatalf("%v\n", err)
	}
	defer nc.Close()
	// NATS

	// SEND Subscribe to ETS
	subMsg := "B1||1|0||"
	ct.SendMsg([]byte(subMsg))

	s.logging.logger.Tracef("SEND " + subMsg)

	var Symbol string
	var Subject string
	for {
		select {
		case msg := <-ct.Incoming:
			// TODO: Processing Here
			m := string(msg)
			s.logging.logger.Tracef("[Msg]", m)
			sl := strings.Split(m, "|")
			Symbol = ""
			Subject = SubjClass
			if sl[0] == "BC" {
				//log.Println("[Symbol]", sl[2])
				Symbol = sl[2]
				if Symbol[0] == '.' {
					Subject += ".INDEX"
				} else {
					Subject += ".STOCK"
				}

				//Symbol = strings.Replace(Symbol, "#", "#0", -1)
				//Symbol = strings.Replace(Symbol, ".", "#1", -1)
				//Symbol = strings.Replace(Symbol, " ", "#2", -1)

				if Symbol[0] == '.' {
					Symbol = Symbol[1:]
				}
				Symbol = strings.Replace(Symbol, ".", "_", -1)
				Symbol = strings.Replace(Symbol, " ", "_", -1)
				//Symbol = url.PathEscape(Symbol)
			}

			if Symbol != "" {
				Subject += "." + Symbol
				err = nc.Publish(Subject, msg)
				if err != nil {
					s.logging.logger.Fatalf("%v\n", err)
				}
				nc.Flush()
				if err := nc.LastError(); err != nil {
					s.logging.logger.Fatalf("%v\n", err)
				} else {
					s.Tracef("Published [%s] : '%s'\n", Subject, msg)
				}
			}
		}
	}

}

// isWindowsService indicates if NATS is running as a Windows service.
func isWindowsService() bool {
	/*
		if dockerized {
			return false
		}
	*/
	isInteractive, _ := svc.IsAnInteractiveSession()
	return !isInteractive
}

// FixBcstApp : Man GORoutine for Windows Service Start
func FixBcstApp(s server) {
	fbs := NewFixBcst()
	fbs.StartUp()
}
