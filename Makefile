include $(GOROOT)/src/Make.inc 

TARG=sealog

GOFILES = \
	config.go \
	log.go \

include $(GOROOT)/src/Make.pkg
