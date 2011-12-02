include $(GOROOT)/src/Make.inc 

TARG=sealog/config
DEPS = ../common ../writers ../dispatchers

GOFILES = \
	logconfig.go \
	parser.go \
	xmlnode.go \

include $(GOROOT)/src/Make.pkg

clean:
	rm -rf *.o *.a *.[$(OS)] [$(OS)].out $(CLEANFILES)
	for i in $(DEPS); do $(MAKE) -C $$i clean; done
	
test:
	gotest
	for i in $(DEPS); do $(MAKE) -C $$i test; done