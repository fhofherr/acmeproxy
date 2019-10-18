package acmeclient

import (
	"fmt"
	"sync"

	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/fhofherr/golf/log"
	"github.com/go-acme/lego/certcrypto"
	legolog "github.com/go-acme/lego/log"
	"github.com/pkg/errors"
)

var legoOnce sync.Once

// InitializeLego initializes the lego library.
//
// Lego uses a global variable to store a logger. In order to use a different
// loger this variable has to be set. The InitializeLego function uses a
// sync.Once that the Logger is set only once, no matter how often the function
// is called. However, direct access to leoglog.Logger cannot be protected.
func InitializeLego(logger log.Logger) {
	legoOnce.Do(func() {
		legolog.Logger = &loggerAdapter{
			Logger: logger,
		}
	})
}

type loggerAdapter struct {
	Logger log.Logger
}

func (l *loggerAdapter) Fatal(args ...interface{}) {
	log.Log(l.Logger, "level", "error", "message", fmt.Sprint(args...))
}

func (l *loggerAdapter) Fatalln(args ...interface{}) {
	log.Log(l.Logger, "level", "error", "message", fmt.Sprint(args...))
}

func (l *loggerAdapter) Fatalf(format string, args ...interface{}) {
	log.Log(l.Logger, "level", "error", "message", fmt.Sprintf(format, args...))
}

func (l *loggerAdapter) Print(args ...interface{}) {
	log.Log(l.Logger, "level", "info", "message", fmt.Sprint(args...))
}

func (l *loggerAdapter) Println(args ...interface{}) {
	log.Log(l.Logger, "level", "info", "message", fmt.Sprint(args...))
}

func (l *loggerAdapter) Printf(format string, args ...interface{}) {
	log.Log(l.Logger, "level", "info", "message", fmt.Sprintf(format, args...))
}

func legoKeyType(kt certutil.KeyType) (certcrypto.KeyType, error) {
	switch kt {
	case certutil.EC256:
		return certcrypto.EC256, nil
	case certutil.EC384:
		return certcrypto.EC384, nil
	case certutil.RSA2048:
		return certcrypto.RSA2048, nil
	case certutil.RSA4096:
		return certcrypto.RSA4096, nil
	case certutil.RSA8192:
		return certcrypto.RSA8192, nil
	default:
		return "", errors.Errorf("unsupported key type: %v", kt)
	}
}
