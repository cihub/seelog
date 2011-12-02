include $(GOROOT)/src/Make.inc 

TARG=github.com/cihub/sealog/writers
DEPS = ../test

GOFILES = \
	fswrapper.go \
	filewriter.go \
	bufferedwriter.go \
	consolewriter.go \
	rollingfilewriter.go \

include $(GOROOT)/src/Make.pkg

clean:

	rm -rf *.o *.a *.[$(OS)] [$(OS)].out $(CLEANFILES)
	for i in $(DEPS); do $(MAKE) -C $$i clean; done
	
test:


ifdef FULLTEST
	gotest
else
	gotest -test.run=".*Test([^F]|F[^i]|Fi[^l]|Fil[^e]|File[^S]).*"
endif

	for i in $(DEPS); do $(MAKE) -C $$i test; done
