include $(GOROOT)/src/Make.inc 

TARG=../../testlibrary
DEPS = ../../..

GOFILES = \
	lib.go \

include $(GOROOT)/src/Make.pkg

clean:
	rm -rf *.o *.a *.[$(OS)] [$(OS)].out $(CLEANFILES)
	for i in $(DEPS); do $(MAKE) -C $$i clean; done
	
test:
	for i in $(DEPS); do $(MAKE) -C $$i test; done