all:
	@echo "Available targets are: tag-patch"
	@echo "Run 'make tag-patch' to create and push a new patch version."


# Related to tag-patch
TAG_PREFIX := v
DEFAULT_VERSION := 0.0.0

LATEST_TAG := $(shell git describe --tags --abbrev=0 --match "$(TAG_PREFIX)*" 2>/dev/null || echo $(TAG_PREFIX)$(DEFAULT_VERSION))
CURRENT_VERSION := $(shell echo $(LATEST_TAG) | sed "s/^$(TAG_PREFIX)//")
MAJOR := $(shell echo $(CURRENT_VERSION) | cut -d'.' -f1)
MINOR := $(shell echo $(CURRENT_VERSION) | cut -d'.' -f2)
PATCH := $(shell echo $(CURRENT_VERSION) | cut -d'.' -f3)
NEXT_PATCH := $(shell echo $$(($(PATCH) + 1)))
NEW_VERSION := $(MAJOR).$(MINOR).$(NEXT_PATCH)
NEW_TAG := $(TAG_PREFIX)$(NEW_VERSION)

tag-patch:
	@echo "Latest existing tag: $(LATEST_TAG)"
	@echo "Next Patch Version:  $(NEW_TAG)"
	@git tag -a $(NEW_TAG) -m "Release $(NEW_TAG)"
	@read -p "Do you want to push tag $(NEW_TAG) to origin? (y/N): " CONFIRM; \
	if [ "$$CONFIRM" = "y" ] || [ "$$CONFIRM" = "Y" ]; then \
		git push origin $(NEW_TAG); \
	else \
		echo "Removing local tag $(NEW_TAG)..."; \
		git tag -d $(NEW_TAG); \
	fi
