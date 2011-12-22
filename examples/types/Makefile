include $(GOROOT)/src/Make.inc

TARG=types
DEPS=../..
GOFILES=\
	main.go \

include $(GOROOT)/src/Make.cmd

clean:
	rm -rf *.o *.a *.[$(OS)] [$(OS)].out $(CLEANFILES)
	for i in $(DEPS); do $(MAKE) -C $$i clean; done
	
test:
	for i in $(DEPS); do $(MAKE) -C $$i test; done
