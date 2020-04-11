package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"fmt"
	"net"

	"github.com/mhale/smtpd"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ttys3/smtp-brd/provider"
	"github.com/ttys3/smtp-brd/parser"
)

var appName = "smtp-brd"
var Version = "dev"

var listenAddr string
var listenPort string
var tlsEn bool
var certFile string
var keyFile string

//@TODO support real AUTH
var authUsername string
var authPassword string

var curProvider = ""
var curSender provider.Sender
var showVersion bool
var showHelp bool
var debugEn bool
var logger *zap.Logger

func init() {
	viper.SetConfigName("config.toml") // name of config file
	viper.AddConfigPath(".")               // optionally look for config in the working directory
	viper.AddConfigPath("/etc/" + appName)   // path to look for the config file in
	viper.SetEnvPrefix("SMTP_BRD")
	viper.AutomaticEnv()

	viper.Set("Verbose", true)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; just warn
			fmt.Println("config file not found.")
		} else {
			// Config file was found but another error was produced
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
	}

	flag.BoolVarP(&showVersion, "version", "v", false, "show app version")
	flag.BoolVarP(&showHelp, "help", "h", false, "show help")
	flag.BoolVarP(&debugEn, "debug", "d", false, "enable debug")
	flag.StringVarP(&curProvider, "provider", "P", "mailgun", "enable email send service provider")
	flag.StringVarP(&listenAddr, "listen", "l", "0.0.0.0", "listen address")
	flag.StringVarP(&listenPort, "port", "p", "2525", "listen port")
	flag.BoolVarP(&tlsEn, "tls", "t", false, "enable TLS")
	flag.StringVarP(&certFile, "cert", "c", "", "TLS certificate file path")
	flag.StringVarP(&keyFile, "key", "k", "", "TLS private key file path")
	flag.StringVarP(&authUsername, "user", "u", "", "SMTP AUTH username")
	flag.StringVarP(&authPassword, "secret", "s", "", "SMTP AUTH password")

	flag.Parse()

	if err := viper.BindPFlags(flag.CommandLine); err != nil {
		panic(fmt.Errorf("viper.BindPFlags err: %w", err))
	}
}

func mailHandler(origin net.Addr, from string, to []string, data []byte) {
	// skip all Received header
	msg , err := parser.ParseMail(data)
	if err != nil {
		zap.S().Errorf("parse mail failed: %s, Received mail from %s for %s with subject %s, origin: %s",
			err, from, to[0], msg.Subject, origin.String())
		return
	}
	zap.S().Infof("Received mail from %s for %s with subject %s, origin: %s\n plainMessage: %s\n htmlMessage: %s\n, attachments: %#v",
		from, to[0], msg.Subject, origin.String(),
		msg.BodyPlain, msg.BodyHtml, msg.Attachments,
		)
	if err := curSender.Send(from, to[0], msg.Subject, string(msg.BodyPlain), string(msg.BodyHtml)); err != nil {
		zap.S().Errorf("mail send failed: %s", err)
	}
}

func main() {
	flushLog := initZapLogger(debugEn)
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
	if tlsEn && (certFile == "" || keyFile == "") {
		zap.S().Fatalf("TLS can not be enabled without specific cert and key path")
	}
	if listenAddr == "" {
		zap.S().Fatalf("listen addr can not be empty")
	}
	if listenPort == "" {
		zap.S().Fatalf("listen port can not be empty")
	}
	addr :=  listenAddr + ":" + listenPort
	zap.S().Infof("server listen on %s", addr)
	if tlsEn {
		zap.S().Info("TLS enabled")
	}

	if authUsername != "" && authPassword != "" {
		zap.S().Info("SMTP AUTH enabled")
	}

	if curProvider == "" {
		zap.S().Fatalf("provider can not be empty")
	}
	if factory, err := provider.GetFactory(curProvider); err != nil {
		zap.S().Fatalf("provider init err: %s", err)
	} else {
		curSender = factory()
		zap.S().Infof("provider init success: %s", curSender)
	}
	var err error
	srv := &smtpd.Server{Addr: addr, Handler: mailHandler, Appname: appName, Hostname: ""}
	srv.AuthHandler = smtpdAuth
	// RFC 4954 specifies that plaintext authentication mechanisms such as LOGIN and PLAIN require a TLS connection.
	// This can be explicitly overridden e.g. setting s.srv.AuthMechs["LOGIN"] = true.
	// warn: if you disabled TLS, the go smtp client will only work if the hostname is "localhost", "127.0.0.1" or "::1"
	// see https://golang.org/src/net/smtp/auth.go#L46
	srv.AuthMechs = map[string]bool{"LOGIN": true, "PLAIN": true, "CRAM-MD5": true}
	if debugEn {
		smtpd.Debug = true
		srv.LogRead = smtpdLogger("read")
		srv.LogWrite = smtpdLogger("write")
	}
	if tlsEn {
		err = srv.ConfigureTLS(certFile, keyFile)
		if err != nil {
			zap.S().Fatalf("TLS server start failed with error: %s", err)
		}
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
	//zapCfg := zap.NewProductionConfig()
	//zapCfg.Encoding = "console"
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
	logger = logger.WithOptions(zap.AddCallerSkip(1))
	//The default global logger used by zap.L() and zap.S() is a no-op logger.
	//To configure the global loggers, you must use ReplaceGlobals.
	zap.ReplaceGlobals(logger)
	if dbg {
		zap.S().Infof("debug enabled")
	}
	return func() {
		// flushes buffer, if any
		if err := logger.Sync(); err != nil {
			fmt.Printf("zap: Sync() failed with error: %s\n", err)
		}
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
func smtpdAuth(remoteAddr net.Addr, mechanism string, username []byte, password []byte, sharedKey []byte) (bool, error) {
	zap.S().Debugf("[smtp.AuthHandler] remoteAddr: %s, mechanism: %s, got username: [%s], password: [%s], sharedKey: [%s]",
		remoteAddr, mechanism, username, password, sharedKey)
	// skip auth if the server does not require
	if authUsername == "" || authPassword == "" {
		return true, nil
	}
	errAuth := fmt.Errorf("invalid username or password")
	if bytes.Compare(username, []byte(authUsername)) != 0 {
		zap.S().Debugf("username expect: %s, actual: %s", authUsername, username)
		// username invalid
		return false, errAuth
	}
	if mechanism == "CRAM-MD5" {
		d := hmac.New(md5.New, []byte(authPassword))
		d.Write(sharedKey)
		s := make([]byte, 0, d.Size())
		expectPwdHmac := []byte(fmt.Sprintf("%x", d.Sum(s)))
		// password invalid
		if bytes.Compare(password, expectPwdHmac) != 0 {
			zap.S().Debugf("password expect: %s, actual: %s", expectPwdHmac, password)
			return false, errAuth
		}
	} else {
		// password invalid
		if bytes.Compare(password, []byte(authPassword)) != 0 {
			zap.S().Debugf("password expect: %s, actual: %s", authPassword, password)
			return false, errAuth
		}
	}
	return true, nil
}