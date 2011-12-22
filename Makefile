include $(GOROOT)/src/Make.inc 

TARG=github.com/cihub/sealog
DEPS=common config

GOFILES=\
	config.go \
	logger.go \
	synclogger.go \
	asynclogger.go \
	asynclooplogger.go \
	asynctimerlogger.go \
	log.go \

include $(GOROOT)/src/Make.pkg

clean:
	rm -rf *.o *.a *.[$(OS)] [$(OS)].out $(CLEANFILES)
	for i in $(DEPS); do $(MAKE) -C $$i clean; done
	
test:
	gotest
	for i in $(DEPS); do $(MAKE) -C $$i test; done
