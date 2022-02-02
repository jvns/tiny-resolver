package main

import (
	"flag"
	"os"
	"strings"

	"github.com/jvns/resolve/logger"
	"github.com/jvns/resolve/resolver"
	"go.uber.org/zap"
)

func main() {
	var (
		target     *string
		nameserver *string
	)

	target = flag.String("target", "", "desired record for resolving")
	nameserver = flag.String("nameserver", "", "desired nameserver used for resolving")

	flag.Parse()

	if *target == "" {
		logger.Log.Error("Target is required")
		os.Exit(1)
	}

	if !strings.HasSuffix(*target, ".") {
		*target = *target + "."
	}

	log := logger.Log.WithOptions(zap.Fields(
		zap.String("target", *target),
	))

	r := resolver.New(nameserver, nil, log)
	ip, err := r.Resolve(*target)

	if err != nil {
		log.Error("Error in resolving",
			zap.Error(err),
		)
	} else {
		log.Info("Got result",
			zap.String("ip", ip.String()),
		)
	}
}
