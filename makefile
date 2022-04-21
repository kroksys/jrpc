# Version upgrade
APP_VERSION=$(shell cat version)
VERSION_MAJOR=$(shell echo $(APP_VERSION) | cut -d. -f1)
VERSION_MINOR=$(shell echo $(APP_VERSION) | cut -d. -f2)
VERSION_MICRO=$(shell echo $(APP_VERSION) | cut -d. -f3)
VERSION_MICRO_NEXT=$(shell echo $$(($(VERSION_MICRO)+1)))
VERSION_NEXT=$(shell echo "$(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_MICRO_NEXT)")
upgrade:
	@echo $(VERSION_NEXT) > version
	git tag $(shell cat version)
	git push origin --tags
.PHONY: upgrade