include $(GOROOT)/src/Make.inc 

TARG=sealog/dispatchers
DEPS = ../common ../format

GOFILES = \
	formattedwriter.go \
	dispatcher.go \
	splitdispatcher.go \
	filterdispatcher.go \

include $(GOROOT)/src/Make.pkg

clean:
	rm -rf *.o *.a *.[$(OS)] [$(OS)].out $(CLEANFILES)
	for i in $(DEPS); do $(MAKE) -C $$i clean; done

test:
	gotest
	for i in $(DEPS); do $(MAKE) -C $$i test; done