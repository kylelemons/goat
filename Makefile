include $(GOROOT)/src/Make.inc

TARG=goat
GOFILES=*.go

DEPS=term

include $(GOROOT)/src/Make.cmd

demo: all
	@echo "Try it out!"
	@echo " 1. Type lines and see them echoed"
	@echo " 2. Press UP and edit your previous line"
	@echo " 3. Press DOWN to get to the end of the line"
	@./goat;
