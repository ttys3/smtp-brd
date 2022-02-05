package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"fmt"
	"net"
	"os"

	flag "github.com/spf13/pflag"
	"github.com/ttys3/smtpd"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ttys3/smtp-brd/config"
	"github.com/ttys3/smtp-brd/parser"
	"github.com/ttys3/smtp-brd/provider"
)

var (
	appName = "smtp-brd"
	Version = "dev"
)

var (
	logger      *zap.Logger
	sndrFactory provider.Factory
)

var (
	configFile  string
	showVersion bool
	showHelp    bool
	dumpConfig  bool
	debug       bool
)

func init() {
	flag.StringVarP(&configFile, "config", "c", "", "config file path")
	flag.BoolVarP(&showVersion, "version", "v", false, "show app version")
	flag.BoolVarP(&showHelp, "help", "h", false, "show help")
	flag.BoolVarP(&debug, "debug", "d", false, "enable debug")
	flag.BoolVarP(&dumpConfig, "dump", "", false, "dump demo config")
	flag.CommandLine.MarkHidden("debug")

	// cfg := *config.Cfg()
	// flag.StringVarP(&cfg.Provider, "provider", "P", "mailgun", "enable email send service provider")
	// flag.StringVarP(&cfg.Addr, "addr", "l", "0.0.0.0", "listen address")
	// flag.StringVarP(&cfg.Port, "port", "p", "2525", "listen port")
	// flag.StringVarP(&cfg.AuthUsername,"user", "u", "", "SMTP AUTH username")
	// flag.StringVarP(&cfg.AuthPassword, "secret", "s", "", "SMTP AUTH password")
}

func initCfg() {
	// do this after all providers has been registered
	flag.Parse()

	if dumpConfig {
		config.Cfg().Dump()
		os.Exit(0)
	}

	err := config.Load(configFile)
	if err != nil {
		zap.S().Fatal(err)
	}
}

func mailHandler(origin net.Addr, from string, to []string, data []byte) error {
	msg, err := parser.ParseMail(data)
	if err != nil {
		e := fmt.Errorf("parse mail failed with err: %s, received mail from %s to %s with subject %s, request origin: %s",
			err, from, to[0], msg.Subject, origin.String())
		zap.S().Error(e)
		return e
	}
	zap.S().Debugf("parse mail success, received mail from %s for %s with subject %s, "+
		"request origin: %s\nmsg parsed: %#v",
		from, to[0], msg.Subject, origin.String(), msg)
	sndr := sndrFactory()
	sndr.AddCCs(msg.CC...)
	sndr.AddBCCs(msg.BCC...)
	sndr.AddTos(msg.To...)
	if err := sndr.Send(from, "", msg.Subject, string(msg.BodyPlain), string(msg.BodyHtml)); err != nil {
		e := fmt.Errorf("mail send failed with err: %s, from: %s, to: %s, subject: %s, request origin: %s", err,
			from, to[0], msg.Subject, origin.String())
		zap.S().Error(e)
		return e
	}
	return nil
}

func main() {
	initCfg()
	flushLog := initZapLogger(config.Cfg().Debug)
	defer flushLog()

	if showVersion {
		fmt.Println("Name:    smtp-brd")
		fmt.Println("Version: " + Version)
		fmt.Println("Author:  荒野無燈")
		return
	}
	if showHelp {
		flag.Usage()
		return
	}

	zap.S().Infof("config: %+v", config.Cfg())

	if config.Cfg().TLS && (config.Cfg().CertFile == "" || config.Cfg().KeyFile == "") {
		zap.S().Fatalf("TLS can not be enabled without specific cert and key path")
	}
	if config.Cfg().Addr == "" {
		zap.S().Fatalf("listen addr can not be empty")
	}
	if config.Cfg().Port == "" {
		zap.S().Fatalf("listen port can not be empty")
	}
	addr := config.Cfg().Addr + ":" + config.Cfg().Port
	zap.S().Infof("server listen on %s", addr)

	if config.Cfg().AuthUsername != "" && config.Cfg().AuthPassword != "" {
		zap.S().Info("SMTP AUTH enabled")
	}

	zap.S().Infof("available providers: %s", provider.AvailableProviders())

	if config.Cfg().Provider == "" {
		zap.S().Fatalf("provider can not be empty")
	}
	if factory, err := provider.GetFactory(config.Cfg().Provider); err != nil {
		zap.S().Fatalf("provider init err: %s", err)
	} else {
		sndrFactory = factory
		zap.S().Infof("provider init success: [%s]", config.Cfg().Provider)
	}
	// start the server
	initSmtpd(addr, appName, "")
}

func initSmtpd(addr, appName, hostname string) {
	var err error
	srv := &smtpd.Server{Addr: addr, Handler: mailHandler, Appname: appName, Hostname: hostname}
	srv.AuthHandler = smtpdAuth
	if config.Cfg().AuthUsername != "" && config.Cfg().AuthPassword != "" {
		srv.AuthRequired = true
	}
	// RFC 4954 specifies that plaintext authentication mechanisms such as LOGIN and PLAIN require a TLS connection.
	// This can be explicitly overridden e.g. setting s.srv.AuthMechs["LOGIN"] = true.
	// warn: if you disabled TLS, the go smtp client will only work if the hostname is "localhost", "127.0.0.1" or "::1"
	// see https://golang.org/src/net/smtp/auth.go#L46
	srv.AuthMechs = map[string]bool{"LOGIN": true, "PLAIN": true, "CRAM-MD5": true}
	if config.Cfg().Debug {
		smtpd.Debug = true
		srv.LogRead = smtpdLogger("read")
		srv.LogWrite = smtpdLogger("write")
	}
	if config.Cfg().TLS {
		err = srv.ConfigureTLS(config.Cfg().CertFile, config.Cfg().KeyFile)
		if err != nil {
			zap.S().Fatalf("TLS server start failed with error: %s", err)
		} else {
			// golang net/smtp client tls.Config.InsecureSkipVerify is enforced to true
			// if TLS enabled, be sure your have valid certificate
			// self-signed cert is not allowed by golang smtp client
			zap.S().Infof("TLS server started. cert file: %s, key file: %s", config.Cfg().CertFile, config.Cfg().KeyFile)
		}
		srv.TLSListener = true
		srv.TLSRequired = true
	}
	err = srv.ListenAndServe()
	if err != nil {
		zap.S().Errorf("server exited with error: %s", err)
	}
}

func initZapLogger(dbg bool) func() {
	// prod stackLevel := ErrorLevel
	// dev stackLevel = WarnLevel
	zapCfg := zap.NewDevelopmentConfig()
	// zapCfg := zap.NewProductionConfig()
	// zapCfg.Encoding = "console"
	if !dbg {
		zapCfg.DisableCaller = true
		zapCfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	} else {
		zapCfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}
	// if Development , stackLevel = WarnLevel
	zapCfg.Development = false
	zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	tmpLogger, _ := zapCfg.Build()
	logger = tmpLogger.Named("[" + appName + "]")
	// logger = logger.WithOptions(zap.AddCallerSkip(1))
	// The default global logger used by zap.L() and zap.S() is a no-op logger.
	// To configure the global loggers, you must use ReplaceGlobals.
	zap.ReplaceGlobals(logger)
	if dbg {
		zap.S().Infof("debug enabled")
	}
	return func() {
		if logger == nil {
			return
		}
		// flushes buffer, if any
		logger.Sync()
	}
}

func smtpdLogger(logType string) smtpd.LogFunc {
	return func(remoteIP, verb, line string) {
		zap.S().Infof("[smtpd.%s] remoteIP: %s, ver: %s, line: %s", logType, remoteIP, verb, line)
	}
}

// SMTP server AUTH handler
// see http://www.ietf.org/rfc/rfc4616.txt
//
// AUTH PLAIN
// With AUTH PLAIN, the credentials should be sent according to this grammar
// [authzid] UTF8NUL authcid UTF8NUL passwd
// authzid usually is empty
// echo -ne "\0username\0password"| base64
// AUTH LOGIN
// the server ask the client send username and password separately, both are base64 encoded
// AUTH CRAM-MD5
// bash telnet test, username=admin, password=secret
// decode the challenge-response from server:
// SERVER_SEND_KEY=$(echo -ne 'PDE3MTQzMDkuMTk3MzMxMTkzQHNtdHAtYnJkPg==' | base64 -d)
// calc the hmac md5 hash:
// HMAC_HASH=$(echo -ne ${SERVER_SEND_KEY} | openssl dgst -md5 -hmac "secret")
// got the final message:
// echo -ne "admin ${HMAC_HASH}" | base64
func smtpdAuth(remoteAddr net.Addr, mechanism string, username []byte, password []byte, challenge []byte) (bool, error) {
	zap.S().Debugf("[smtp.AuthHandler] remoteAddr: %s, mechanism: %s, got username: [%s], password: [%s], challenge: [%s]",
		remoteAddr, mechanism, username, password, challenge)
	// skip auth if the server does not require
	if config.Cfg().AuthUsername == "" || config.Cfg().AuthPassword == "" {
		return true, nil
	}
	errAuth := fmt.Errorf("invalid username or password")
	if !bytes.Equal(username, []byte(config.Cfg().AuthUsername)) {
		zap.S().Debugf("username expect: %s, actual: %s", config.Cfg().AuthUsername, username)
		// username invalid
		return false, errAuth
	}
	if mechanism == "CRAM-MD5" {
		d := hmac.New(md5.New, []byte(config.Cfg().AuthPassword))
		d.Write(challenge)
		s := make([]byte, 0, d.Size())
		expectPwdHmac := []byte(fmt.Sprintf("%x", d.Sum(s)))
		// password invalid
		if !bytes.Equal(password, expectPwdHmac) {
			zap.S().Debugf("password expect: %s, actual: %s", expectPwdHmac, password)
			return false, errAuth
		}
	} else {
		// AUTH LOGIN/PLAIN
		if !bytes.Equal(password, []byte(config.Cfg().AuthPassword)) {
			zap.S().Debugf("password expect: %s, actual: %s", config.Cfg().AuthPassword, password)
			return false, errAuth
		}
	}
	return true, nil
}
