include $(GOROOT)/src/Make.inc 

TARG=sealog

GOFILES = \
	constraints.go \
	exception.go \
	config.go \
	log.go \

include $(GOROOT)/src/Make.pkg
